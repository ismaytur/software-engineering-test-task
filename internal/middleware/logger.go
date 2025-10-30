package middleware

import (
	"log/slog"
	"time"

	"cruder/pkg/logger"

	"github.com/gin-gonic/gin"
)

const requestLoggerKey = "request.logger"

func RequestLogger(base *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		reqLogger := base.With(
			slog.String("http.request.method", c.Request.Method),
			slog.String("http.request.path", c.Request.URL.Path),
			slog.String("http.request.host", c.Request.Host),
			slog.String("http.client_ip", c.ClientIP()),
		)

		if route := c.FullPath(); route != "" {
			reqLogger = reqLogger.With(slog.String("http.route", route))
		}
		if rid := c.GetHeader("X-Request-ID"); rid != "" {
			reqLogger = reqLogger.With(slog.String("http.request.id", rid))
		}

		c.Set(requestLoggerKey, reqLogger)
		ctx := logger.ContextWithLogger(c.Request.Context(), reqLogger)
		c.Request = c.Request.WithContext(ctx)

		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()

		attrs := []any{
			slog.Int("http.response.status_code", status),
			slog.Duration("http.server.request.duration", duration),
		}

		if len(c.Errors) > 0 {
			reqLogger.Error("request completed with errors",
				append(attrs, slog.String("error", c.Errors.String()))...)
			return
		}

		reqLogger.Info("request handled", attrs...)
	}
}

func LoggerFromContext(c *gin.Context, fallback *logger.Logger) *logger.Logger {
	if l, exists := c.Get(requestLoggerKey); exists {
		if reqLogger, ok := l.(*logger.Logger); ok && reqLogger != nil {
			return reqLogger
		}
	}
	return fallback
}
