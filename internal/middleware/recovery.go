package middleware

import (
	"net/http"
	"runtime/debug"

	"log/slog"

	"cruder/pkg/logger"

	"github.com/gin-gonic/gin"
)

func Recovery(log *logger.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		reqLogger := LoggerFromContext(c, log).With(
			slog.Any("panic", recovered),
			slog.String("stacktrace", string(debug.Stack())),
		)
		reqLogger.Error("panic recovered")
		c.AbortWithStatus(http.StatusInternalServerError)
	})
}
