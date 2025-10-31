package service

import (
	"cruder/internal/repository"
	"time"
)

type Service struct {
	Users   UserService
	APIKeys APIKeyService
}

func NewService(repos *repository.Repository, apiKeyTTL time.Duration) *Service {
	return &Service{
		Users:   NewUserService(repos.Users),
		APIKeys: NewAPIKeyService(repos.APIKeys, apiKeyTTL),
	}
}
