package service

//go:generate go run github.com/vektra/mockery/v2@latest --config=../../mockery.yaml

import (
	"errors"
	"testing"

	"cruder/internal/model"
	"cruder/internal/repository"
	"cruder/internal/service/mocks"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var errUnexpected = errors.New("unexpected error")

func TestUserService_Create_Success(t *testing.T) {
	// Given: a repository that accepts user creation
	repo := mocks.NewUserRepositoryMock(t)
	service := NewUserService(repo)
	repo.On("Create", "new_user", "user@example.com", "Test User").
		Return(&model.User{
			ID:       1,
			UUID:     uuid.NewString(),
			Username: "new_user",
			Email:    "user@example.com",
			FullName: "Test User",
		}, nil).Once()

	// When: creating a user with padded fields
	user, err := service.Create("  new_user  ", "user@example.com", "  Test User ")

	// Then: the user is created and trimmed input was passed to the repository
	require.NoError(t, err)
	require.NotNil(t, user)
	require.Equal(t, "new_user", user.Username)
	repo.AssertExpectations(t)
}

func TestUserService_Create_InvalidEmail(t *testing.T) {
	// Given: a user service with a mock repository
	repo := mocks.NewUserRepositoryMock(t)
	service := NewUserService(repo)

	// When: creating a user with malformed email
	_, err := service.Create("name", "invalid-email", "Full Name")

	// Then: invalid user input error is returned
	require.ErrorIs(t, err, ErrInvalidUserInput)
	repo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)
}

func TestUserService_Create_Duplicate(t *testing.T) {
	// Given: repository returns unique violation
	repo := mocks.NewUserRepositoryMock(t)
	service := NewUserService(repo)
	repo.On("Create", "dup_user", "dup@example.com", "Dup User").
		Return((*model.User)(nil), repository.ErrUniqueViolation).Once()

	// When: creating a user with duplicate data
	_, err := service.Create("dup_user", "dup@example.com", "Dup User")

	// Then: duplicate error is translated to ErrUserAlreadyExists
	require.ErrorIs(t, err, ErrUserAlreadyExists)
	repo.AssertExpectations(t)
}

func TestUserService_GetAll_Success(t *testing.T) {
	repo := mocks.NewUserRepositoryMock(t)
	service := NewUserService(repo)
	expected := []model.User{{ID: 1}, {ID: 2}}
	repo.On("GetAll").Return(expected, nil).Once()

	users, err := service.GetAll()

	require.NoError(t, err)
	require.Equal(t, expected, users)
	repo.AssertExpectations(t)
}

func TestUserService_GetAll_Error(t *testing.T) {
	repo := mocks.NewUserRepositoryMock(t)
	service := NewUserService(repo)
	repo.On("GetAll").Return(nil, errUnexpected).Once()

	users, err := service.GetAll()

	require.Error(t, err)
	require.Nil(t, users)
}

func TestUserService_UpdateByUUID_Success(t *testing.T) {
	// Given: repository has an existing user and accepts update
	existing := &model.User{
		ID:       10,
		UUID:     uuid.NewString(),
		Username: "current",
		Email:    "current@example.com",
		FullName: "Current Name",
	}

	repo := mocks.NewUserRepositoryMock(t)
	service := NewUserService(repo)
	repo.On("GetByUUID", mock.AnythingOfType("uuid.UUID")).Return(existing, nil).Once()
	repo.On("UpdateByUUID", mock.AnythingOfType("uuid.UUID"), "current", "current@example.com", "Updated Name").
		Return(&model.User{
			ID:       existing.ID,
			UUID:     existing.UUID,
			Username: "current",
			Email:    "current@example.com",
			FullName: "Updated Name",
		}, nil).Once()
	newName := "  Updated Name "

	// When: updating only the full name
	result, err := service.UpdateByUUID(uuid.MustParse(existing.UUID), UpdateUserInput{
		FullName: strPtr(newName),
	})

	// Then: repository receives merged values and result reflects changes
	require.NoError(t, err)
	require.Equal(t, "Updated Name", result.FullName)
	repo.AssertExpectations(t)
}

func TestUserService_UpdateByUUID_InvalidEmail(t *testing.T) {
	// Given: an existing user in repository
	existing := &model.User{
		ID:       10,
		UUID:     uuid.NewString(),
		Username: "current",
		Email:    "current@example.com",
		FullName: "Current Name",
	}
	repo := mocks.NewUserRepositoryMock(t)
	service := NewUserService(repo)
	repo.On("GetByUUID", mock.AnythingOfType("uuid.UUID")).Return(existing, nil).Once()
	badEmail := "not-an-email"

	// When: updating with an invalid email value
	_, err := service.UpdateByUUID(uuid.MustParse(existing.UUID), UpdateUserInput{
		Email: &badEmail,
	})

	// Then: invalid user input error is returned
	require.ErrorIs(t, err, ErrInvalidUserInput)
	repo.AssertExpectations(t)
}

func TestUserService_UpdateByUUID_NoFieldsProvided(t *testing.T) {
	// Given: user service with a mock repository
	repo := mocks.NewUserRepositoryMock(t)
	service := NewUserService(repo)

	// When: updating without providing any fields
	_, err := service.UpdateByUUID(uuid.New(), UpdateUserInput{})

	// Then: invalid user input error is returned
	require.ErrorIs(t, err, ErrInvalidUserInput)
	repo.AssertNotCalled(t, "GetByUUID", mock.Anything)
}

func TestUserService_GetByUsername_Success(t *testing.T) {
	repo := mocks.NewUserRepositoryMock(t)
	service := NewUserService(repo)
	existing := &model.User{Username: "tester"}
	repo.On("GetByUsername", "tester").Return(existing, nil).Once()

	user, err := service.GetByUsername("tester")

	require.NoError(t, err)
	require.Equal(t, existing, user)
	repo.AssertExpectations(t)
}

func TestUserService_GetByUsername_NotFound(t *testing.T) {
	repo := mocks.NewUserRepositoryMock(t)
	service := NewUserService(repo)
	repo.On("GetByUsername", "missing").Return((*model.User)(nil), nil).Once()

	user, err := service.GetByUsername("missing")

	require.ErrorIs(t, err, ErrUserNotFound)
	require.Nil(t, user)
}

func TestUserService_GetByUsername_Error(t *testing.T) {
	repo := mocks.NewUserRepositoryMock(t)
	service := NewUserService(repo)
	repo.On("GetByUsername", "err").Return((*model.User)(nil), errUnexpected).Once()

	user, err := service.GetByUsername("err")

	require.Error(t, err)
	require.Nil(t, user)
}

func TestUserService_GetByID_Success(t *testing.T) {
	repo := mocks.NewUserRepositoryMock(t)
	service := NewUserService(repo)
	existing := &model.User{ID: 10}
	repo.On("GetByID", int64(10)).Return(existing, nil).Once()

	user, err := service.GetByID(10)

	require.NoError(t, err)
	require.Equal(t, existing, user)
	repo.AssertExpectations(t)
}

func TestUserService_GetByID_NotFound(t *testing.T) {
	repo := mocks.NewUserRepositoryMock(t)
	service := NewUserService(repo)
	repo.On("GetByID", int64(11)).Return((*model.User)(nil), nil).Once()

	user, err := service.GetByID(11)

	require.ErrorIs(t, err, ErrUserNotFound)
	require.Nil(t, user)
}

func TestUserService_GetByID_Error(t *testing.T) {
	repo := mocks.NewUserRepositoryMock(t)
	service := NewUserService(repo)
	repo.On("GetByID", int64(12)).Return((*model.User)(nil), errUnexpected).Once()

	user, err := service.GetByID(12)

	require.Error(t, err)
	require.Nil(t, user)
}

func TestUserService_GetByUUID_Success(t *testing.T) {
	repo := mocks.NewUserRepositoryMock(t)
	service := NewUserService(repo)
	u := uuid.New()
	existing := &model.User{UUID: u.String()}
	repo.On("GetByUUID", u).Return(existing, nil).Once()

	user, err := service.GetByUUID(u)

	require.NoError(t, err)
	require.Equal(t, existing, user)
	repo.AssertExpectations(t)
}

func TestUserService_GetByUUID_NotFound(t *testing.T) {
	repo := mocks.NewUserRepositoryMock(t)
	service := NewUserService(repo)
	u := uuid.New()
	repo.On("GetByUUID", u).Return((*model.User)(nil), nil).Once()

	user, err := service.GetByUUID(u)

	require.ErrorIs(t, err, ErrUserNotFound)
	require.Nil(t, user)
}

func TestUserService_GetByUUID_Error(t *testing.T) {
	repo := mocks.NewUserRepositoryMock(t)
	service := NewUserService(repo)
	u := uuid.New()
	repo.On("GetByUUID", u).Return((*model.User)(nil), errUnexpected).Once()

	user, err := service.GetByUUID(u)

	require.Error(t, err)
	require.Nil(t, user)
}

func TestUserService_UpdateByUUID_DuplicateEmail(t *testing.T) {
	// Given: repository returns unique violation on update
	existing := &model.User{
		ID:       10,
		UUID:     uuid.NewString(),
		Username: "current",
		Email:    "current@example.com",
		FullName: "Current Name",
	}
	repo := mocks.NewUserRepositoryMock(t)
	service := NewUserService(repo)
	repo.On("GetByUUID", mock.AnythingOfType("uuid.UUID")).Return(existing, nil).Once()
	repo.On("UpdateByUUID", mock.AnythingOfType("uuid.UUID"), "current", mock.Anything, mock.Anything).
		Return((*model.User)(nil), repository.ErrUniqueViolation).Once()
	newEmail := "duplicate@example.com"

	// When: updating email that conflicts with existing user
	_, err := service.UpdateByUUID(uuid.MustParse(existing.UUID), UpdateUserInput{
		Email: &newEmail,
	})

	// Then: ErrUserAlreadyExists is returned
	require.ErrorIs(t, err, ErrUserAlreadyExists)
	repo.AssertExpectations(t)
}

func TestUserService_DeleteByUUID_Success(t *testing.T) {
	// Given: repository successfully deletes a user
	repo := mocks.NewUserRepositoryMock(t)
	service := NewUserService(repo)
	repo.On("DeleteByUUID", mock.AnythingOfType("uuid.UUID")).Return(true, nil).Once()

	// When: deleting an existing user
	err := service.DeleteByUUID(uuid.New())

	// Then: no error is returned
	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestUserService_DeleteByUUID_NotFound(t *testing.T) {
	// Given: repository reports user not found
	repo := mocks.NewUserRepositoryMock(t)
	service := NewUserService(repo)
	repo.On("DeleteByUUID", mock.AnythingOfType("uuid.UUID")).Return(false, nil).Once()

	// When: deleting a non-existent user
	err := service.DeleteByUUID(uuid.New())

	// Then: ErrUserNotFound is returned
	require.ErrorIs(t, err, ErrUserNotFound)
	repo.AssertExpectations(t)
}

func TestUserService_UpdateByID_InvalidID(t *testing.T) {
	// Given: user service with mock repository
	repo := mocks.NewUserRepositoryMock(t)
	service := NewUserService(repo)

	// When: updating using an invalid (non-positive) ID
	_, err := service.UpdateByID(0, UpdateUserInput{
		FullName: strPtr("Name"),
	})

	// Then: invalid user input error is returned
	require.ErrorIs(t, err, ErrInvalidUserInput)
	repo.AssertNotCalled(t, "GetByID", mock.Anything)
}

func TestUserService_DeleteByID_InvalidID(t *testing.T) {
	repo := mocks.NewUserRepositoryMock(t)
	service := NewUserService(repo)

	err := service.DeleteByID(0)

	require.ErrorIs(t, err, ErrInvalidUserInput)
	repo.AssertNotCalled(t, "DeleteByID", mock.Anything)
}

func TestUserService_DeleteByID_Success(t *testing.T) {
	repo := mocks.NewUserRepositoryMock(t)
	service := NewUserService(repo)
	repo.On("DeleteByID", int64(15)).Return(true, nil).Once()

	err := service.DeleteByID(15)

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestUserService_DeleteByID_NotFound(t *testing.T) {
	repo := mocks.NewUserRepositoryMock(t)
	service := NewUserService(repo)
	repo.On("DeleteByID", int64(16)).Return(false, nil).Once()

	err := service.DeleteByID(16)

	require.ErrorIs(t, err, ErrUserNotFound)
	repo.AssertExpectations(t)
}

func TestUserService_UpdateByID_EmailValidation(t *testing.T) {
	// Given: repository contains an existing user
	existing := &model.User{
		ID:       42,
		UUID:     uuid.NewString(),
		Username: "current",
		Email:    "current@example.com",
		FullName: "Holder",
	}
	repo := mocks.NewUserRepositoryMock(t)
	service := NewUserService(repo)
	newEmail := "updated@example.com"
	repo.On("GetByID", int64(existing.ID)).Return(existing, nil).Once()
	repo.On("UpdateByID", int64(existing.ID), "current", newEmail, "Holder").
		Return(&model.User{
			ID:       existing.ID,
			UUID:     existing.UUID,
			Username: "current",
			Email:    newEmail,
			FullName: "Holder",
		}, nil).Once()

	// When: updating email to a valid address
	result, err := service.UpdateByID(int64(existing.ID), UpdateUserInput{
		Email: &newEmail,
	})

	// Then: repository receives new email and returns updated user
	require.NoError(t, err)
	require.Equal(t, newEmail, result.Email)
	repo.AssertExpectations(t)
}

// strPtr returns a pointer to the provided string.
func strPtr(s string) *string {
	return &s
}
