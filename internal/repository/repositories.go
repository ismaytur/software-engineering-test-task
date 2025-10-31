package repository

import "database/sql"

type Repository struct {
	Users   UserRepository
	APIKeys APIKeyRepository
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		Users:   NewUserRepository(db),
		APIKeys: NewAPIKeyRepository(db),
	}
}
