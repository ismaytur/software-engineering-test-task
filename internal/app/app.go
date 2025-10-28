package app

import (
	"fmt"

	"cruder/internal/controller"
	"cruder/internal/handler"
	"cruder/internal/repository"
	"cruder/internal/service"

	"github.com/gin-gonic/gin"
)

type App struct {
	Engine  *gin.Engine
	Service *service.Service

	conn repository.DatabaseConnection
}

func New(dsn string) (*App, error) {
	if dsn == "" {
		return nil, fmt.Errorf("DSN cannot be empty")
	}

	dbConn, err := repository.NewPostgresConnection(dsn)
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	repos := repository.NewRepository(dbConn.DB())
	services := service.NewService(repos)
	controllers := controller.NewController(services)

	router := gin.New()
	router.Use(gin.Recovery())
	handler.New(router, controllers.Users)

	return &App{
		Engine:  router,
		Service: services,
		conn:    dbConn,
	}, nil
}

func (a *App) Close() error {
	if a == nil || a.conn == nil {
		return nil
	}
	return a.conn.DB().Close()
}
