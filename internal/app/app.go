package app

import (
	"fmt"
	"log/slog"

	"cruder/internal/controller"
	"cruder/internal/handler"
	"cruder/internal/middleware"
	"cruder/internal/repository"
	"cruder/internal/service"
	"cruder/pkg/logger"

	"github.com/gin-gonic/gin"
)

type App struct {
	Engine  *gin.Engine
	Service *service.Service

	Logger *logger.Logger

	conn repository.DatabaseConnection
}

func New(dsn string) (*App, error) {
	if dsn == "" {
		return nil, fmt.Errorf("DSN cannot be empty")
	}

	baseLogger := logger.Get()
	appLogger := baseLogger.With(slog.String("component", "app"))
	appLogger.Info("initializing application")

	gin.DefaultWriter = logger.Writer(baseLogger, slog.LevelInfo)
	gin.DefaultErrorWriter = logger.Writer(baseLogger, slog.LevelError)

	appLogger.Info("connecting to database")
	dbConn, err := repository.NewPostgresConnection(dsn)
	if err != nil {
		appLogger.Error("failed to connect to database", slog.String("error", err.Error()))
		return nil, fmt.Errorf("connect to database: %w", err)
	}
	appLogger.Info("database connection established")

	repos := repository.NewRepository(dbConn.DB())
	services := service.NewService(repos)
	controllers := controller.NewController(services)

	router := gin.New()
	router.Use(
		middleware.Recovery(appLogger),
		middleware.RequestLogger(appLogger),
	)
	handler.New(router, controllers.Users)
	appLogger.Info("http router configured")

	return &App{
		Engine:  router,
		Service: services,
		Logger:  appLogger,
		conn:    dbConn,
	}, nil
}

func (a *App) Close() error {
	if a == nil || a.conn == nil {
		return nil
	}

	a.Logger.Info("closing database connection")
	return a.conn.DB().Close()
}
