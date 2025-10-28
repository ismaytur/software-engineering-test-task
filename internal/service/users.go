package service

import (
	"cruder/internal/model"
	"cruder/internal/repository"
	"errors"
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
}

type UpdateUserInput struct {
	Username *string
	Email    *string
	FullName *string
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) GetAll() ([]model.User, error) {
	users, err := s.repo.GetAll()
	if err != nil {
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
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *userService) GetByID(id int64) (*model.User, error) {
	user, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *userService) GetByUUID(uuid uuid.UUID) (*model.User, error) {
	user, err := s.repo.GetByUUID(uuid)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *userService) Create(username, email, fullName string) (*model.User, error) {
	username = strings.TrimSpace(username)
	email = strings.TrimSpace(email)
	fullName = strings.TrimSpace(fullName)

	if username == "" || fullName == "" {
		return nil, ErrInvalidUserInput
	}

	_, err := mail.ParseAddress(email)
	if err != nil {
		return nil, ErrInvalidUserInput
	}

	user, err := s.repo.Create(username, email, fullName)
	if err != nil {
		if errors.Is(err, repository.ErrUniqueViolation) {
			return nil, ErrUserAlreadyExists
		}
		return nil, err
	}

	return user, nil
}

func (s *userService) UpdateByUUID(uuid uuid.UUID, input UpdateUserInput) (*model.User, error) {
	if input.Username == nil && input.Email == nil && input.FullName == nil {
		return nil, ErrInvalidUserInput
	}

	existing, err := s.repo.GetByUUID(uuid)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrUserNotFound
	}

	username := existing.Username
	email := existing.Email
	fullName := existing.FullName

	if input.Username != nil {
		trimmed := strings.TrimSpace(*input.Username)
		if trimmed == "" {
			return nil, ErrInvalidUserInput
		}
		username = trimmed
	}

	if input.Email != nil {
		trimmed := strings.TrimSpace(*input.Email)
		if trimmed == "" {
			return nil, ErrInvalidUserInput
		}
		if _, err := mail.ParseAddress(trimmed); err != nil {
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
			return nil, ErrUserAlreadyExists
		}
		return nil, err
	}
	if updated == nil {
		return nil, ErrUserNotFound
	}
	return updated, nil
}

func (s *userService) DeleteByUUID(uuid uuid.UUID) error {
	ok, err := s.repo.DeleteByUUID(uuid)
	if err != nil {
		return err
	}
	if !ok {
		return ErrUserNotFound
	}
	return nil
}

func (s *userService) UpdateByID(id int64, input UpdateUserInput) (*model.User, error) {
	if id <= 0 {
		return nil, ErrInvalidUserInput
	}

	if input.Username == nil && input.Email == nil && input.FullName == nil {
		return nil, ErrInvalidUserInput
	}

	existing, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrUserNotFound
	}

	username := existing.Username
	email := existing.Email
	fullName := existing.FullName

	if input.Username != nil {
		trimmed := strings.TrimSpace(*input.Username)
		if trimmed == "" {
			return nil, ErrInvalidUserInput
		}
		username = trimmed
	}

	if input.Email != nil {
		trimmed := strings.TrimSpace(*input.Email)
		if trimmed == "" {
			return nil, ErrInvalidUserInput
		}
		if _, err := mail.ParseAddress(trimmed); err != nil {
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
			return nil, ErrUserAlreadyExists
		}
		return nil, err
	}
	if updated == nil {
		return nil, ErrUserNotFound
	}
	return updated, nil
}

func (s *userService) DeleteByID(id int64) error {
	if id <= 0 {
		return ErrInvalidUserInput
	}

	ok, err := s.repo.DeleteByID(id)
	if err != nil {
		return err
	}
	if !ok {
		return ErrUserNotFound
	}
	return nil
}
