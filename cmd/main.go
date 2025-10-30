package main

import (
	"fmt"
	"log/slog"
	"os"

	"cruder/internal/app"
	"cruder/pkg/logger"
)

func main() {
	envOptions := map[string]string{
		"LOG_OUTPUT": os.Getenv("LOG_OUTPUT"),
		"LOG_FILE":   os.Getenv("LOG_FILE"),
		"LOG_LEVEL":  os.Getenv("LOG_LEVEL"),
	}
	logOptions := logger.OptionsFromEnv(envOptions)

	appLogger, err := logger.Configure(logOptions)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to configure logger: %v\n", err)
		os.Exit(1)
	}

	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		appLogger.Error("no postgres DSN is defined, exiting")
		os.Exit(1)
	}

	application, err := app.New(dsn)
	if err != nil {
		appLogger.Error("failed to initialize application", slog.String("error", err.Error()))
		os.Exit(1)
	}
	appLogger.Info("application initialized")
	defer func() {
		if err := application.Close(); err != nil {
			appLogger.Warn("failed to close application cleanly", slog.String("error", err.Error()))
		}
		if appLogger != nil {
			if err := appLogger.Close(); err != nil {
				fmt.Fprintf(os.Stderr, "failed to close logger: %v\n", err)
			}
		}
	}()

	appLogger.Info("starting http server")
	if err := application.Engine.Run(); err != nil {
		appLogger.Error("failed to run server", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
