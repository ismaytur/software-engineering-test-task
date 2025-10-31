package service

import (
	"context"
	"cruder/internal/model"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAPIKeyServiceValidate(t *testing.T) {
	repo := newMockAPIKeyRepository()
	svc := NewAPIKeyService(repo, time.Minute)

	ctx := context.Background()

	_, err := svc.Validate(ctx, "")
	require.ErrorIs(t, err, ErrAPIKeyMissing)

	_, err = svc.Validate(ctx, "missing")
	require.ErrorIs(t, err, ErrAPIKeyInvalid)
	require.Equal(t, 1, repo.callCount(hashAPIKey("missing")))

	_, err = svc.Validate(ctx, "missing")
	require.ErrorIs(t, err, ErrAPIKeyInvalid)
	require.Equal(t, 2, repo.callCount(hashAPIKey("missing")), "negative results are not cached")

	key, err := svc.Validate(ctx, "valid-key")
	require.NoError(t, err)
	require.Equal(t, "Test Client", key.ClientName)
	require.Equal(t, 1, repo.callCount(hashAPIKey("valid-key")))

	// cache hit should avoid second repo call
	_, err = svc.Validate(ctx, "valid-key")
	require.NoError(t, err)
	require.Equal(t, 1, repo.callCount(hashAPIKey("valid-key")))
}

type mockAPIKeyRepository struct {
	data  map[string]*model.APIKey
	calls map[string]int
}

func newMockAPIKeyRepository() *mockAPIKeyRepository {
	validHash := hashAPIKey("valid-key")
	return &mockAPIKeyRepository{
		data: map[string]*model.APIKey{
			validHash: {
				ID:         1,
				KeyHash:    validHash,
				ClientName: "Test Client",
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
		},
		calls: make(map[string]int),
	}
}

func (m *mockAPIKeyRepository) GetByHash(_ context.Context, hash string) (*model.APIKey, error) {
	m.calls[hash]++
	key, ok := m.data[hash]
	if !ok {
		return nil, nil
	}
	return key, nil
}

func (m *mockAPIKeyRepository) callCount(hash string) int {
	return m.calls[hash]
}
