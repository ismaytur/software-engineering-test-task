package controller

import (
	"errors"
	"net/http"

	"cruder/internal/controller/request"
	"cruder/internal/controller/response"
	"cruder/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	errInvalidID   = "invalid id"
	errInvalidUUID = "invalid uuid"
	errInvalidBody = "invalid payload"
)

type UserController struct {
	service service.UserService
}

func NewUserController(service service.UserService) *UserController {
	return &UserController{service: service}
}

// GetAllUsers godoc
// @Summary      List users
// @Tags         users
// @Produce      json
// @Success      200  {array}   response.User
// @Failure      500  {object}  response.Error
// @Router       /api/v1/users/ [get]
func (c *UserController) GetAllUsers(ctx *gin.Context) {
	users, err := c.service.GetAll()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, response.Error{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, users)
}

// GetUserByUsername godoc
// @Summary      Fetch user by username
// @Tags         users
// @Param        username  path      string  true  "User username"
// @Produce      json
// @Success      200  {object}  response.User
// @Failure      404  {object}  response.Error
// @Failure      500  {object}  response.Error
// @Router       /api/v1/users/username/{username} [get]
func (c *UserController) GetUserByUsername(ctx *gin.Context) {
	username := ctx.Param("username")

	user, err := c.service.GetByUsername(username)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			ctx.JSON(http.StatusNotFound, response.Error{Error: err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, response.Error{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, user)
}

// GetUserByID godoc
// @Summary      Fetch user by ID
// @Tags         users
// @Param        id   path      int  true  "User ID"
// @Produce      json
// @Success      200  {object}  response.User
// @Failure      400  {object}  response.Error
// @Failure      404  {object}  response.Error
// @Failure      500  {object}  response.Error
// @Router       /api/v1/users/id/{id} [get]
func (c *UserController) GetUserByID(ctx *gin.Context) {
	var uri request.IDParam
	if err := ctx.ShouldBindUri(&uri); err != nil {
		ctx.JSON(http.StatusBadRequest, response.Error{Error: errInvalidID})
		return
	}

	user, err := c.service.GetByID(uri.ID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			ctx.JSON(http.StatusNotFound, response.Error{Error: err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, response.Error{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, user)
}

// GetUserByUUID godoc
// @Summary      Fetch user by UUID
// @Tags         users
// @Param        uuid  path      string  true  "User UUID"
// @Produce      json
// @Success      200  {object}  response.User
// @Failure      400  {object}  response.Error
// @Failure      404  {object}  response.Error
// @Failure      500  {object}  response.Error
// @Router       /api/v1/users/uuid/{uuid} [get]
func (c *UserController) GetUserByUUID(ctx *gin.Context) {
	var uri request.UUIDParam
	if err := ctx.ShouldBindUri(&uri); err != nil {
		ctx.JSON(http.StatusBadRequest, response.Error{Error: errInvalidUUID})
		return
	}

	parsedUUID, err := uuid.Parse(uri.UUID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, response.Error{Error: errInvalidUUID})
		return
	}

	user, err := c.service.GetByUUID(parsedUUID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			ctx.JSON(http.StatusNotFound, response.Error{Error: err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, response.Error{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, user)
}

// CreateUser godoc
// @Summary      Create user
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request  body      request.CreateUser  true  "User payload"
// @Success      201  {object}  response.User
// @Failure      400  {object}  response.Error
// @Failure      409  {object}  response.Error
// @Failure      500  {object}  response.Error
// @Router       /api/v1/users/ [post]
func (c *UserController) CreateUser(ctx *gin.Context) {
	var req request.CreateUser
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, response.Error{Error: errInvalidBody})
		return
	}

	user, err := c.service.Create(req.Username, req.Email, req.FullName)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidUserInput):
			ctx.JSON(http.StatusBadRequest, response.Error{Error: err.Error()})
			return
		case errors.Is(err, service.ErrUserAlreadyExists):
			ctx.JSON(http.StatusConflict, response.Error{Error: err.Error()})
			return
		default:
			ctx.JSON(http.StatusInternalServerError, response.Error{Error: err.Error()})
			return
		}
	}

	ctx.JSON(http.StatusCreated, user)
}

// UpdateUserByUUID godoc
// @Summary      Update user by UUID
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        uuid     path      string             true  "User UUID"
// @Param        request  body      request.UpdateUser  true  "User payload"
// @Success      200  {object}  response.User
// @Failure      400  {object}  response.Error
// @Failure      404  {object}  response.Error
// @Failure      409  {object}  response.Error
// @Failure      500  {object}  response.Error
// @Router       /api/v1/users/uuid/{uuid} [patch]
func (c *UserController) UpdateUserByUUID(ctx *gin.Context) {
	var uri request.UUIDParam
	if err := ctx.ShouldBindUri(&uri); err != nil {
		ctx.JSON(http.StatusBadRequest, response.Error{Error: errInvalidUUID})
		return
	}

	parsedUUID, err := uuid.Parse(uri.UUID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, response.Error{Error: errInvalidUUID})
		return
	}

	var req request.UpdateUser
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, response.Error{Error: errInvalidBody})
		return
	}

	updated, err := c.service.UpdateByUUID(parsedUUID, service.UpdateUserInput{
		Username: req.Username,
		Email:    req.Email,
		FullName: req.FullName,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidUserInput):
			ctx.JSON(http.StatusBadRequest, response.Error{Error: err.Error()})
			return
		case errors.Is(err, service.ErrUserNotFound):
			ctx.JSON(http.StatusNotFound, response.Error{Error: err.Error()})
			return
		case errors.Is(err, service.ErrUserAlreadyExists):
			ctx.JSON(http.StatusConflict, response.Error{Error: err.Error()})
			return
		default:
			ctx.JSON(http.StatusInternalServerError, response.Error{Error: err.Error()})
			return
		}
	}

	ctx.JSON(http.StatusOK, updated)
}

// DeleteUserByUUID godoc
// @Summary      Delete user by UUID
// @Tags         users
// @Param        uuid  path  string  true  "User UUID"
// @Success      204  "No Content"
// @Failure      400  {object}  response.Error
// @Failure      404  {object}  response.Error
// @Failure      500  {object}  response.Error
// @Router       /api/v1/users/uuid/{uuid} [delete]
func (c *UserController) DeleteUserByUUID(ctx *gin.Context) {
	var uri request.UUIDParam
	if err := ctx.ShouldBindUri(&uri); err != nil {
		ctx.JSON(http.StatusBadRequest, response.Error{Error: errInvalidUUID})
		return
	}

	parsedUUID, err := uuid.Parse(uri.UUID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, response.Error{Error: errInvalidUUID})
		return
	}

	if err := c.service.DeleteByUUID(parsedUUID); err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidUserInput):
			ctx.JSON(http.StatusBadRequest, response.Error{Error: err.Error()})
			return
		case errors.Is(err, service.ErrUserNotFound):
			ctx.JSON(http.StatusNotFound, response.Error{Error: err.Error()})
			return
		default:
			ctx.JSON(http.StatusInternalServerError, response.Error{Error: err.Error()})
			return
		}
	}

	ctx.Status(http.StatusNoContent)
}

// UpdateUserByID godoc
// @Summary      Update user by ID
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id       path      int               true  "User ID"
// @Param        request  body      request.UpdateUser  true  "User payload"
// @Success      200  {object}  response.User
// @Failure      400  {object}  response.Error
// @Failure      404  {object}  response.Error
// @Failure      409  {object}  response.Error
// @Failure      500  {object}  response.Error
// @Router       /api/v1/users/id/{id} [patch]
func (c *UserController) UpdateUserByID(ctx *gin.Context) {
	var uri request.IDParam
	if err := ctx.ShouldBindUri(&uri); err != nil {
		ctx.JSON(http.StatusBadRequest, response.Error{Error: errInvalidID})
		return
	}

	var req request.UpdateUser
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, response.Error{Error: errInvalidBody})
		return
	}

	updated, err := c.service.UpdateByID(uri.ID, service.UpdateUserInput{
		Username: req.Username,
		Email:    req.Email,
		FullName: req.FullName,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidUserInput):
			ctx.JSON(http.StatusBadRequest, response.Error{Error: err.Error()})
			return
		case errors.Is(err, service.ErrUserNotFound):
			ctx.JSON(http.StatusNotFound, response.Error{Error: err.Error()})
			return
		case errors.Is(err, service.ErrUserAlreadyExists):
			ctx.JSON(http.StatusConflict, response.Error{Error: err.Error()})
			return
		default:
			ctx.JSON(http.StatusInternalServerError, response.Error{Error: err.Error()})
			return
		}
	}

	ctx.JSON(http.StatusOK, updated)
}

// DeleteUserByID godoc
// @Summary      Delete user by ID
// @Tags         users
// @Param        id  path  int  true  "User ID"
// @Success      204  "No Content"
// @Failure      400  {object}  response.Error
// @Failure      404  {object}  response.Error
// @Failure      500  {object}  response.Error
// @Router       /api/v1/users/id/{id} [delete]
func (c *UserController) DeleteUserByID(ctx *gin.Context) {
	var uri request.IDParam
	if err := ctx.ShouldBindUri(&uri); err != nil {
		ctx.JSON(http.StatusBadRequest, response.Error{Error: errInvalidID})
		return
	}

	if err := c.service.DeleteByID(uri.ID); err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidUserInput):
			ctx.JSON(http.StatusBadRequest, response.Error{Error: err.Error()})
			return
		case errors.Is(err, service.ErrUserNotFound):
			ctx.JSON(http.StatusNotFound, response.Error{Error: err.Error()})
			return
		default:
			ctx.JSON(http.StatusInternalServerError, response.Error{Error: err.Error()})
			return
		}
	}

	ctx.Status(http.StatusNoContent)
}
