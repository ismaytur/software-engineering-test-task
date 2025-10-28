package service

import (
	"cruder/internal/model"
	"cruder/internal/repository"
	"errors"
)

var ErrUserNotFound = errors.New("user not found")

type UserService interface {
	GetAll() ([]model.User, error)
	GetByUsername(username string) (*model.User, error)
	GetByID(id int64) (*model.User, error)
}

type userService struct {
	repo repository.UserRepository
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
