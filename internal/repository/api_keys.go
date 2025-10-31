package repository

import (
	"context"
	"cruder/internal/model"
	"database/sql"
	"errors"
)

type APIKeyRepository interface {
	GetByHash(ctx context.Context, hash string) (*model.APIKey, error)
}

type apiKeyRepository struct {
	db *sql.DB
}

func NewAPIKeyRepository(db *sql.DB) APIKeyRepository {
	return &apiKeyRepository{db: db}
}

func (r *apiKeyRepository) GetByHash(ctx context.Context, hash string) (*model.APIKey, error) {
	var key model.APIKey
	err := r.db.QueryRowContext(
		ctx,
		`SELECT id, key_hash, client_name, created_at, updated_at FROM api_keys WHERE key_hash = $1`,
		hash,
	).Scan(&key.ID, &key.KeyHash, &key.ClientName, &key.CreatedAt, &key.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &key, nil
}
