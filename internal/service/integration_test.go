//go:build integration

package service_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"cruder/internal/app"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/pressly/goose/v3"
)

var (
	testDB     *sql.DB
	testServer *httptest.Server
	testApp    *app.App
	apiBaseURL string
	dsn        string
)

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("could not connect to docker: %v", err)
	}

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "18-alpine",
		Env: []string{
			"POSTGRES_USER=postgres",
			"POSTGRES_PASSWORD=postgres",
			"POSTGRES_DB=postgres",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		log.Fatalf("could not start postgres container: %v", err)
	}

	port := resource.GetPort("5432/tcp")
	dsn = fmt.Sprintf("host=localhost port=%s user=postgres password=postgres dbname=postgres sslmode=disable", port)

	if err := pool.Retry(func() error {
		var err error
		testDB, err = sql.Open("postgres", dsn)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		return testDB.PingContext(ctx)
	}); err != nil {
		log.Fatalf("could not connect to postgres: %v", err)
	}

	if err := runMigrations(testDB); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	gin.SetMode(gin.TestMode)
	testApp, err = app.New(dsn)
	if err != nil {
		log.Fatalf("failed to initialize application: %v", err)
	}

	testServer = httptest.NewServer(testApp.Engine)
	apiBaseURL = testServer.URL

	code := m.Run()

	testServer.Close()
	if err := testApp.Close(); err != nil {
		log.Printf("failed to close application: %v", err)
	}
	_ = testDB.Close()
	if err := pool.Purge(resource); err != nil {
		log.Printf("failed to purge resource: %v", err)
	}
	os.Exit(code)
}

func resetUsersTable(t *testing.T) {
	t.Helper()
	if _, err := testDB.Exec("TRUNCATE users RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate users: %v", err)
	}
	if err := seedUsers(); err != nil {
		t.Fatalf("failed to seed users: %v", err)
	}
}

func seedUsers() error {
	rows := []struct {
		username string
		email    string
		fullName string
	}{
		{"jdoe", "jdoe@example.com", "John Doe"},
		{"asmith", "asmith@example.com", "Alice Smith"},
		{"bjones", "bjones@example.com", "Bob Jones"},
	}
	for _, row := range rows {
		if _, err := testDB.Exec(
			`INSERT INTO users (username, email, full_name) VALUES ($1, $2, $3)`,
			row.username, row.email, row.fullName,
		); err != nil {
			return err
		}
	}
	return nil
}

func runMigrations(db *sql.DB) error {
	root, err := projectRoot()
	if err != nil {
		return err
	}

	migrationsDir := filepath.Join(root, "migrations")
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}

	if err := goose.Up(db, migrationsDir); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}
	return nil
}

func projectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("go.mod not found")
		}
		dir = parent
	}
}
