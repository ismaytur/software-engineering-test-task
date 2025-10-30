package service

import (
	"cruder/internal/model"
	"cruder/internal/repository"
	"cruder/pkg/logger"
	"errors"
	"log/slog"
	"net/mail"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidUserInput  = errors.New("invalid user input")
	ErrUserAlreadyExists = errors.New("user already exists")
)

type UserService interface {
	GetAll() ([]model.User, error)
	GetByUsername(username string) (*model.User, error)
	GetByID(id int64) (*model.User, error)
	GetByUUID(uuid uuid.UUID) (*model.User, error)
	Create(username, email, fullName string) (*model.User, error)
	UpdateByUUID(uuid uuid.UUID, input UpdateUserInput) (*model.User, error)
	DeleteByUUID(uuid uuid.UUID) error
	UpdateByID(id int64, input UpdateUserInput) (*model.User, error)
	DeleteByID(id int64) error
}

type userService struct {
	repo repository.UserRepository
	log  *logger.Logger
}

type UpdateUserInput struct {
	Username *string
	Email    *string
	FullName *string
}

func NewUserService(repo repository.UserRepository) UserService {
	serviceLogger := logger.Get().With(slog.String("component", "service.user"))
	return &userService{
		repo: repo,
		log:  serviceLogger,
	}
}

func (s *userService) GetAll() ([]model.User, error) {
	users, err := s.repo.GetAll()
	if err != nil {
		s.log.Error("failed to fetch users", slog.String("error", err.Error()))
		return nil, err
	}
	if users == nil {
		return []model.User{}, nil
	}
	return users, nil
}

func (s *userService) GetByUsername(username string) (*model.User, error) {
	user, err := s.repo.GetByUsername(username)
	if err != nil {
		s.log.Error("failed to fetch user by username", slog.String("user.username", username), slog.String("error", err.Error()))
		return nil, err
	}
	if user == nil {
		s.log.Debug("user by username not found", slog.String("user.username", username))
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *userService) GetByID(id int64) (*model.User, error) {
	user, err := s.repo.GetByID(id)
	if err != nil {
		s.log.Error("failed to fetch user by id", slog.Int64("user.id", id), slog.String("error", err.Error()))
		return nil, err
	}
	if user == nil {
		s.log.Debug("user by id not found", slog.Int64("user.id", id))
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *userService) GetByUUID(uuid uuid.UUID) (*model.User, error) {
	user, err := s.repo.GetByUUID(uuid)
	if err != nil {
		s.log.Error("failed to fetch user by uuid", slog.String("user.uuid", uuid.String()), slog.String("error", err.Error()))
		return nil, err
	}
	if user == nil {
		s.log.Debug("user by uuid not found", slog.String("user.uuid", uuid.String()))
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *userService) Create(username, email, fullName string) (*model.User, error) {
	username = strings.TrimSpace(username)
	email = strings.TrimSpace(email)
	fullName = strings.TrimSpace(fullName)

	if username == "" || fullName == "" {
		s.log.Warn("create user invalid input: missing username or full name")
		return nil, ErrInvalidUserInput
	}

	_, err := mail.ParseAddress(email)
	if err != nil {
		s.log.Warn("create user invalid email format")
		return nil, ErrInvalidUserInput
	}

	user, err := s.repo.Create(username, email, fullName)
	if err != nil {
		if errors.Is(err, repository.ErrUniqueViolation) {
			s.log.Warn("create user duplicate", slog.String("user.username", username))
			return nil, ErrUserAlreadyExists
		}
		s.log.Error("create user repository error", slog.String("error", err.Error()))
		return nil, err
	}

	s.log.Info("user created", slog.String("user.uuid", user.UUID), slog.Int("user.id", user.ID))
	return user, nil
}

func (s *userService) UpdateByUUID(uuid uuid.UUID, input UpdateUserInput) (*model.User, error) {
	if input.Username == nil && input.Email == nil && input.FullName == nil {
		s.log.Warn("update by uuid invalid input: no fields provided", slog.String("user.uuid", uuid.String()))
		return nil, ErrInvalidUserInput
	}

	existing, err := s.repo.GetByUUID(uuid)
	if err != nil {
		s.log.Error("failed to fetch existing user by uuid", slog.String("user.uuid", uuid.String()), slog.String("error", err.Error()))
		return nil, err
	}
	if existing == nil {
		s.log.Warn("update by uuid target not found", slog.String("user.uuid", uuid.String()))
		return nil, ErrUserNotFound
	}

	username := existing.Username
	email := existing.Email
	fullName := existing.FullName

	if input.Username != nil {
		trimmed := strings.TrimSpace(*input.Username)
		if trimmed == "" {
			s.log.Warn("update by uuid invalid username", slog.String("user.uuid", uuid.String()))
			return nil, ErrInvalidUserInput
		}
		username = trimmed
	}

	if input.Email != nil {
		trimmed := strings.TrimSpace(*input.Email)
		if trimmed == "" {
			s.log.Warn("update by uuid empty email", slog.String("user.uuid", uuid.String()))
			return nil, ErrInvalidUserInput
		}
		if _, err := mail.ParseAddress(trimmed); err != nil {
			s.log.Warn("update by uuid invalid email", slog.String("user.uuid", uuid.String()))
			return nil, ErrInvalidUserInput
		}
		email = trimmed
	}

	if input.FullName != nil {
		trimmed := strings.TrimSpace(*input.FullName)
		fullName = trimmed
	}

	updated, err := s.repo.UpdateByUUID(uuid, username, email, fullName)
	if err != nil {
		if errors.Is(err, repository.ErrUniqueViolation) {
			s.log.Warn("update by uuid duplicate", slog.String("user.uuid", uuid.String()))
			return nil, ErrUserAlreadyExists
		}
		s.log.Error("update by uuid repository error", slog.String("user.uuid", uuid.String()), slog.String("error", err.Error()))
		return nil, err
	}
	if updated == nil {
		s.log.Warn("update by uuid resulted in not found", slog.String("user.uuid", uuid.String()))
		return nil, ErrUserNotFound
	}
	s.log.Info("user updated by uuid", slog.String("user.uuid", updated.UUID), slog.Int("user.id", updated.ID))
	return updated, nil
}

func (s *userService) DeleteByUUID(uuid uuid.UUID) error {
	ok, err := s.repo.DeleteByUUID(uuid)
	if err != nil {
		s.log.Error("delete by uuid repository error", slog.String("user.uuid", uuid.String()), slog.String("error", err.Error()))
		return err
	}
	if !ok {
		s.log.Warn("delete by uuid target not found", slog.String("user.uuid", uuid.String()))
		return ErrUserNotFound
	}
	s.log.Info("user deleted by uuid", slog.String("user.uuid", uuid.String()))
	return nil
}

func (s *userService) UpdateByID(id int64, input UpdateUserInput) (*model.User, error) {
	if id <= 0 {
		s.log.Warn("update by id invalid id", slog.Int64("user.id", id))
		return nil, ErrInvalidUserInput
	}

	if input.Username == nil && input.Email == nil && input.FullName == nil {
		s.log.Warn("update by id invalid input: no fields provided", slog.Int64("user.id", id))
		return nil, ErrInvalidUserInput
	}

	existing, err := s.repo.GetByID(id)
	if err != nil {
		s.log.Error("failed to fetch existing user by id", slog.Int64("user.id", id), slog.String("error", err.Error()))
		return nil, err
	}
	if existing == nil {
		s.log.Warn("update by id target not found", slog.Int64("user.id", id))
		return nil, ErrUserNotFound
	}

	username := existing.Username
	email := existing.Email
	fullName := existing.FullName

	if input.Username != nil {
		trimmed := strings.TrimSpace(*input.Username)
		if trimmed == "" {
			s.log.Warn("update by id invalid username", slog.Int64("user.id", id))
			return nil, ErrInvalidUserInput
		}
		username = trimmed
	}

	if input.Email != nil {
		trimmed := strings.TrimSpace(*input.Email)
		if trimmed == "" {
			s.log.Warn("update by id empty email", slog.Int64("user.id", id))
			return nil, ErrInvalidUserInput
		}
		if _, err := mail.ParseAddress(trimmed); err != nil {
			s.log.Warn("update by id invalid email", slog.Int64("user.id", id))
			return nil, ErrInvalidUserInput
		}
		email = trimmed
	}

	if input.FullName != nil {
		trimmed := strings.TrimSpace(*input.FullName)
		fullName = trimmed
	}

	updated, err := s.repo.UpdateByID(id, username, email, fullName)
	if err != nil {
		if errors.Is(err, repository.ErrUniqueViolation) {
			s.log.Warn("update by id duplicate", slog.Int64("user.id", id))
			return nil, ErrUserAlreadyExists
		}
		s.log.Error("update by id repository error", slog.Int64("user.id", id), slog.String("error", err.Error()))
		return nil, err
	}
	if updated == nil {
		s.log.Warn("update by id resulted in not found", slog.Int64("user.id", id))
		return nil, ErrUserNotFound
	}
	s.log.Info("user updated by id", slog.Int("user.id", updated.ID), slog.String("user.uuid", updated.UUID))
	return updated, nil
}

func (s *userService) DeleteByID(id int64) error {
	if id <= 0 {
		s.log.Warn("delete by id invalid id", slog.Int64("user.id", id))
		return ErrInvalidUserInput
	}

	ok, err := s.repo.DeleteByID(id)
	if err != nil {
		s.log.Error("delete by id repository error", slog.Int64("user.id", id), slog.String("error", err.Error()))
		return err
	}
	if !ok {
		s.log.Warn("delete by id target not found", slog.Int64("user.id", id))
		return ErrUserNotFound
	}
	s.log.Info("user deleted by id", slog.Int64("user.id", id))
	return nil
}
