package main

import (
	"log"
	"os"

	"cruder/internal/app"
)

func main() {
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		log.Fatal("No postgres DSN is defined, exiting")
		return
	}

	application, err := app.New(dsn)
	if err != nil {
		log.Fatalf("failed to initialize application: %v", err)
	}
	defer func() {
		if err := application.Close(); err != nil {
			log.Printf("failed to close application cleanly: %v", err)
		}
	}()

	if err := application.Engine.Run(); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
