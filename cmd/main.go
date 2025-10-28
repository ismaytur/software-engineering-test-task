package main

import (
	"cruder/internal/controller"
	"cruder/internal/handler"
	"cruder/internal/repository"
	"cruder/internal/service"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		log.Fatal("No postgres DSN is defined, exiting")
		return
	}

	dbConn, err := repository.NewPostgresConnection(dsn)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	repositories := repository.NewRepository(dbConn.DB())
	services := service.NewService(repositories)
	controllers := controller.NewController(services)
	r := gin.Default()
	handler.New(r, controllers.Users)
	if err := r.Run(); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
