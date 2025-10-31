package middleware

import (
	"net/http"

	"cruder/internal/service"
	"cruder/pkg/logger"

	"log/slog"

	"github.com/gin-gonic/gin"
)

const (
	HeaderAPIKey        = "X-API-Key" // #nosec G101: header name only
	ContextAPIClientKey = "api.client"
)

func APIKeyAuth(apiKeys service.APIKeyService, log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader(HeaderAPIKey)
		client, err := apiKeys.Validate(c.Request.Context(), apiKey)
		if err != nil {
			switch err {
			case service.ErrAPIKeyMissing:
				log.Warn("request missing api key", loggerRequestAttrs(c)...)
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing api key"})
				return
			case service.ErrAPIKeyInvalid:
				log.Warn("request with invalid api key", loggerRequestAttrs(c)...)
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "invalid api key"})
				return
			default:
				attrs := append(loggerRequestAttrs(c), slog.String("error", err.Error()))
				log.Error("failed to validate api key", attrs...)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
				return
			}
		}
		if client != nil {
			c.Set(ContextAPIClientKey, client)
			log.Debug("api key accepted", append(loggerRequestAttrs(c), slog.String("client_name", client.ClientName))...)
		}
		c.Next()
	}
}

func loggerRequestAttrs(c *gin.Context) []any {
	route := c.FullPath()
	if route == "" {
		route = c.Request.URL.Path
	}
	return []any{
		slog.String("method", c.Request.Method),
		slog.String("path", route),
		slog.String("remote_addr", c.ClientIP()),
	}
}
