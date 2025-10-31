package middleware

import (
	"context"
	"cruder/internal/model"
	"cruder/internal/service"
	"cruder/pkg/logger"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestAPIKeyAuth_MissingKey(t *testing.T) {
	router, stub := setupAPIKeyRouter(t)
	stub.reset()

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	require.Equal(t, http.StatusUnauthorized, resp.Code)
}

func TestAPIKeyAuth_InvalidKey(t *testing.T) {
	router, stub := setupAPIKeyRouter(t)
	stub.reset()

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set(HeaderAPIKey, "wrong")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	require.Equal(t, http.StatusForbidden, resp.Code)
}

func TestAPIKeyAuth_InternalError(t *testing.T) {
	router, stub := setupAPIKeyRouter(t)
	stub.reset()
	stub.err = errors.New("boom")

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set(HeaderAPIKey, "whatever")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	require.Equal(t, http.StatusInternalServerError, resp.Code)
}

func TestAPIKeyAuth_Success(t *testing.T) {
	router, stub := setupAPIKeyRouter(t)
	stub.reset()
	stub.validKey = "secret"

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set(HeaderAPIKey, "secret")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	require.Equal(t, http.StatusOK, resp.Code)
	require.Contains(t, resp.Body.String(), "Test Client")
}

func setupAPIKeyRouter(t *testing.T) (*gin.Engine, *stubAPIKeyService) {
	gin.SetMode(gin.TestMode)
	_, _ = logger.Configure(logger.DefaultOptions())
	log := logger.Get()

	stub := &stubAPIKeyService{}

	router := gin.New()
	router.Use(APIKeyAuth(stub, log))
	router.GET("/protected", func(c *gin.Context) {
		client, ok := c.Get(ContextAPIClientKey)
		require.True(t, ok)
		apiKey := client.(*model.APIKey)
		c.JSON(http.StatusOK, gin.H{"client": apiKey.ClientName})
	})

	return router, stub
}

type stubAPIKeyService struct {
	validKey string
	err      error
}

func (s *stubAPIKeyService) reset() {
	s.validKey = ""
	s.err = nil
}

func (s *stubAPIKeyService) Validate(_ context.Context, key string) (*model.APIKey, error) {
	if s.err != nil {
		return nil, s.err
	}
	if key == "" {
		return nil, service.ErrAPIKeyMissing
	}
	if key != s.validKey {
		return nil, service.ErrAPIKeyInvalid
	}
	return &model.APIKey{ClientName: "Test Client"}, nil
}
