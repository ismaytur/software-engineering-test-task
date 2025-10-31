//go:build integration

package service_test

import (
	"fmt"
	"net/http"
	"testing"

	"cruder/internal/middleware"
	"cruder/internal/service"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"
)

const usersBasePath = "/api/v1/users"

type userResponse struct {
	ID       int    `json:"id"`
	UUID     string `json:"uuid"`
	Username string `json:"username"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func TestFunctionalUserLifecycle(t *testing.T) {
	resetUsersTable(t)

	// Given: API has seeded users
	var seeded []userResponse
	resp, err := restyClient().R().
		SetResult(&seeded).
		Get(apiBaseURL + usersBasePath + "/")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())
	require.Len(t, seeded, 3)

	// When: creating a new user via HTTP
	payload := map[string]string{
		"username":  "lifecycle_user",
		"email":     "life@example.com",
		"full_name": "Lifecycle User",
	}
	var created userResponse
	resp, err = restyClient().R().
		SetBody(payload).
		SetResult(&created).
		Post(apiBaseURL + usersBasePath + "/")
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode())
	require.Equal(t, payload["username"], created.Username)

	// Then: retrieve by ID
	var fetchedByID userResponse
	resp, err = restyClient().R().
		SetResult(&fetchedByID).
		Get(fmt.Sprintf("%s%s/id/%d", apiBaseURL, usersBasePath, created.ID))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())

	// And: retrieve by username
	var fetchedByUsername userResponse
	resp, err = restyClient().R().
		SetResult(&fetchedByUsername).
		Get(fmt.Sprintf("%s%s/username/%s", apiBaseURL, usersBasePath, created.Username))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())

	// And: delete by UUID
	resp, err = restyClient().R().
		Delete(fmt.Sprintf("%s%s/uuid/%s", apiBaseURL, usersBasePath, created.UUID))
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, resp.StatusCode())

	// And: deleting again yields 404
	var errResp errorResponse
	resp, err = restyClient().R().
		SetError(&errResp).
		Delete(fmt.Sprintf("%s%s/uuid/%s", apiBaseURL, usersBasePath, created.UUID))
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, resp.StatusCode())
	require.Equal(t, service.ErrUserNotFound.Error(), errResp.Error)
}

func TestFunctionalUpdateByUUID(t *testing.T) {
	resetUsersTable(t)
	user := createUser(t, "update_uuid", "update@example.com", "Update UUID")

	updatePayload := map[string]string{
		"username":  "updated_name",
		"email":     "updated@example.com",
		"full_name": "Updated Full Name",
	}

	var updated userResponse
	resp, err := restyClient().R().
		SetBody(updatePayload).
		SetResult(&updated).
		Patch(fmt.Sprintf("%s%s/uuid/%s", apiBaseURL, usersBasePath, user.UUID))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())
	require.Equal(t, updatePayload["username"], updated.Username)
	require.Equal(t, updatePayload["email"], updated.Email)
	require.Equal(t, updatePayload["full_name"], updated.FullName)
}

func TestFunctionalUpdateByUUID_InvalidEmail(t *testing.T) {
	resetUsersTable(t)
	user := createUser(t, "invalid_email", "valid@example.com", "Invalid Email")

	payload := map[string]string{"email": "not-an-email"}
	var errResp errorResponse
	resp, err := restyClient().R().
		SetBody(payload).
		SetError(&errResp).
		Patch(fmt.Sprintf("%s%s/uuid/%s", apiBaseURL, usersBasePath, user.UUID))
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode())
	require.Equal(t, service.ErrInvalidUserInput.Error(), errResp.Error)
}

func TestFunctionalGetUserByUUID(t *testing.T) {
	resetUsersTable(t)
	created := createUser(t, "uuid_lookup", "uuid@example.com", "UUID Lookup")

	// When: retrieving by UUID
	var fetched userResponse
	resp, err := restyClient().R().
		SetResult(&fetched).
		Get(fmt.Sprintf("%s%s/uuid/%s", apiBaseURL, usersBasePath, created.UUID))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())
	require.Equal(t, created.UUID, fetched.UUID)

	// Then: invalid UUID returns 400
	var errResp errorResponse
	resp, err = restyClient().R().
		SetError(&errResp).
		Get(fmt.Sprintf("%s%s/uuid/%s", apiBaseURL, usersBasePath, "not-a-uuid"))
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode())
	require.Equal(t, "invalid uuid", errResp.Error)
}

func TestFunctionalUpdateByID(t *testing.T) {
	resetUsersTable(t)
	user := createUser(t, "update_id", "updateid@example.com", "Update ID")

	newName := map[string]string{"full_name": "Updated Via ID"}
	var updated userResponse
	resp, err := restyClient().R().
		SetBody(newName).
		SetResult(&updated).
		Patch(fmt.Sprintf("%s%s/id/%d", apiBaseURL, usersBasePath, user.ID))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())
	require.Equal(t, newName["full_name"], updated.FullName)
}

func TestFunctionalCreate_Duplicate(t *testing.T) {
	resetUsersTable(t)
	user := createUser(t, "duplicate_user", "dup@example.com", "Dup User")

	payload := map[string]string{
		"username":  user.Username,
		"email":     user.Email,
		"full_name": "Another",
	}
	var errResp errorResponse
	resp, err := restyClient().R().
		SetBody(payload).
		SetError(&errResp).
		Post(apiBaseURL + usersBasePath + "/")
	require.NoError(t, err)
	require.Equal(t, http.StatusConflict, resp.StatusCode())
	require.Equal(t, service.ErrUserAlreadyExists.Error(), errResp.Error)
}

func TestFunctionalGetByUsername_NotFound(t *testing.T) {
	resetUsersTable(t)
	var errResp errorResponse
	resp, err := restyClient().R().
		SetError(&errResp).
		Get(fmt.Sprintf("%s%s/username/%s", apiBaseURL, usersBasePath, "missing_user"))
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, resp.StatusCode())
	require.Equal(t, service.ErrUserNotFound.Error(), errResp.Error)
}

func TestFunctionalDeleteByID(t *testing.T) {
	resetUsersTable(t)
	created := createUser(t, "delete_id", "deleteid@example.com", "Delete ID")

	// When: deleting by numeric ID
	resp, err := restyClient().R().
		Delete(fmt.Sprintf("%s%s/id/%d", apiBaseURL, usersBasePath, created.ID))
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, resp.StatusCode())

	// Then: deleting again returns 404
	var errResp errorResponse
	resp, err = restyClient().R().
		SetError(&errResp).
		Delete(fmt.Sprintf("%s%s/id/%d", apiBaseURL, usersBasePath, created.ID))
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, resp.StatusCode())
	require.Equal(t, service.ErrUserNotFound.Error(), errResp.Error)

	// And: invalid ID gives 400
	resp, err = restyClient().R().
		SetError(&errResp).
		Delete(fmt.Sprintf("%s%s/id/%s", apiBaseURL, usersBasePath, "abc"))
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode())
	require.Equal(t, "invalid id", errResp.Error)
}

func createUser(t *testing.T, username, email, fullName string) userResponse {
	t.Helper()
	payload := map[string]string{
		"username":  username,
		"email":     email,
		"full_name": fullName,
	}
	var user userResponse
	resp, err := restyClient().R().
		SetBody(payload).
		SetResult(&user).
		Post(apiBaseURL + usersBasePath + "/")
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode())

	return user
}

var httpClient = resty.New()

func restyClient() *resty.Client {
	return httpClient.SetBaseURL(apiBaseURL).
		SetHeader(middleware.HeaderAPIKey, testAPIKey)
}
