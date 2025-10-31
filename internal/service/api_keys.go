package service

import (
	"context"
	"cruder/internal/model"
	"cruder/internal/repository"
	"cruder/pkg/logger"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"time"
)

var (
	ErrAPIKeyMissing = errors.New("api key missing")
	ErrAPIKeyInvalid = errors.New("api key invalid")
)

type APIKeyService interface {
	Validate(ctx context.Context, apiKey string) (*model.APIKey, error)
}

type cacheEntry struct {
	key     *model.APIKey
	expires time.Time
}

type apiKeyService struct {
	repo repository.APIKeyRepository
	log  *logger.Logger

	mu    sync.RWMutex
	cache map[string]cacheEntry
	ttl   time.Duration
}

func NewAPIKeyService(repo repository.APIKeyRepository, ttl time.Duration) APIKeyService {
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	serviceLogger := logger.Get().With(slog.String("component", "service.api_key"))
	return &apiKeyService{
		repo:  repo,
		log:   serviceLogger,
		cache: make(map[string]cacheEntry),
		ttl:   ttl,
	}
}

func (s *apiKeyService) Validate(ctx context.Context, apiKey string) (*model.APIKey, error) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		s.log.Warn("missing api key")
		return nil, ErrAPIKeyMissing
	}

	hash := hashAPIKey(apiKey)

	if entry, ok := s.getCached(hash); ok {
		return entry.key, nil
	}

	key, err := s.repo.GetByHash(ctx, hash)
	if err != nil {
		s.log.Error("failed to fetch api key", slog.String("error", err.Error()))
		return nil, err
	}

	if key == nil {
		s.log.Warn("invalid api key provided")
		return nil, ErrAPIKeyInvalid
	}

	entry := cacheEntry{
		key:     key,
		expires: time.Now().Add(s.ttl),
	}
	s.setCache(hash, entry)

	s.log.Debug("api key validated", slog.String("client_name", key.ClientName))
	return key, nil
}

func (s *apiKeyService) getCached(hash string) (cacheEntry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, ok := s.cache[hash]
	if !ok {
		return cacheEntry{}, false
	}
	if time.Now().After(entry.expires) {
		// stale entry, drop on write path
		return cacheEntry{}, false
	}
	return entry, true
}

func (s *apiKeyService) setCache(hash string, entry cacheEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if time.Now().After(entry.expires) {
		delete(s.cache, hash)
		return
	}
	s.cache[hash] = entry
}

func hashAPIKey(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
