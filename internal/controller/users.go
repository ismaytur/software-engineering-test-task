package controller

import (
	"errors"
	"log/slog"
	"net/http"

	"cruder/internal/controller/request"
	"cruder/internal/controller/response"
	"cruder/internal/middleware"
	"cruder/internal/service"
	"cruder/pkg/logger"

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

func (c *UserController) requestLogger(ctx *gin.Context, operation string) *logger.Logger {
	base := middleware.LoggerFromContext(ctx, logger.Get())
	return base.With(
		slog.String("component", "controller.users"),
		slog.String("operation", operation),
	)
}

// GetAllUsers godoc
// @Summary      List users
// @Tags         users
// @Produce      json
// @Success      200  {array}   response.User
// @Failure      500  {object}  response.Error
// @Router       /api/v1/users/ [get]
func (c *UserController) GetAllUsers(ctx *gin.Context) {
	log := c.requestLogger(ctx, "GetAllUsers")

	users, err := c.service.GetAll()
	if err != nil {
		log.Error("failed to fetch users", slog.String("error", err.Error()))
		ctx.JSON(http.StatusInternalServerError, response.Error{Error: err.Error()})
		return
	}

	log.Debug("fetched users", slog.Int("users.count", len(users)))
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
	log := c.requestLogger(ctx, "GetUserByUsername").With(slog.String("request.username", username))

	user, err := c.service.GetByUsername(username)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			log.Warn("user not found")
			ctx.JSON(http.StatusNotFound, response.Error{Error: err.Error()})
			return
		}
		log.Error("failed to fetch user by username", slog.String("error", err.Error()))
		ctx.JSON(http.StatusInternalServerError, response.Error{Error: err.Error()})
		return
	}

	log.Debug("fetched user by username")
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
	log := c.requestLogger(ctx, "GetUserByID")
	var uri request.IDParam
	if err := ctx.ShouldBindUri(&uri); err != nil {
		log.Warn("invalid id parameter", slog.String("error", err.Error()))
		ctx.JSON(http.StatusBadRequest, response.Error{Error: errInvalidID})
		return
	}

	log = log.With(slog.Int64("request.user_id", uri.ID))

	user, err := c.service.GetByID(uri.ID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			log.Warn("user not found")
			ctx.JSON(http.StatusNotFound, response.Error{Error: err.Error()})
			return
		}
		log.Error("failed to fetch user by id", slog.String("error", err.Error()))
		ctx.JSON(http.StatusInternalServerError, response.Error{Error: err.Error()})
		return
	}

	log.Debug("fetched user by id")
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
	log := c.requestLogger(ctx, "GetUserByUUID")
	var uri request.UUIDParam
	if err := ctx.ShouldBindUri(&uri); err != nil {
		log.Warn("invalid uuid parameter", slog.String("error", err.Error()))
		ctx.JSON(http.StatusBadRequest, response.Error{Error: errInvalidUUID})
		return
	}

	parsedUUID, err := uuid.Parse(uri.UUID)
	if err != nil {
		log.Warn("failed to parse uuid", slog.String("request.uuid_raw", uri.UUID))
		ctx.JSON(http.StatusBadRequest, response.Error{Error: errInvalidUUID})
		return
	}

	log = log.With(slog.String("request.user_uuid", parsedUUID.String()))

	user, err := c.service.GetByUUID(parsedUUID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			log.Warn("user not found")
			ctx.JSON(http.StatusNotFound, response.Error{Error: err.Error()})
			return
		}
		log.Error("failed to fetch user by uuid", slog.String("error", err.Error()))
		ctx.JSON(http.StatusInternalServerError, response.Error{Error: err.Error()})
		return
	}

	log.Debug("fetched user by uuid")
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
	log := c.requestLogger(ctx, "CreateUser")
	var req request.CreateUser
	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Warn("invalid request body", slog.String("error", err.Error()))
		ctx.JSON(http.StatusBadRequest, response.Error{Error: errInvalidBody})
		return
	}

	log = log.With(
		slog.String("request.username", req.Username),
		slog.Bool("request.email_provided", req.Email != ""),
		slog.Bool("request.full_name_provided", req.FullName != ""),
	)

	user, err := c.service.Create(req.Username, req.Email, req.FullName)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidUserInput):
			log.Warn("invalid user input", slog.String("error", err.Error()))
			ctx.JSON(http.StatusBadRequest, response.Error{Error: err.Error()})
			return
		case errors.Is(err, service.ErrUserAlreadyExists):
			log.Warn("user already exists", slog.String("error", err.Error()))
			ctx.JSON(http.StatusConflict, response.Error{Error: err.Error()})
			return
		default:
			log.Error("failed to create user", slog.String("error", err.Error()))
			ctx.JSON(http.StatusInternalServerError, response.Error{Error: err.Error()})
			return
		}
	}

	log.Info("user created", slog.String("user.uuid", user.UUID), slog.Int("user.id", user.ID))
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
	log := c.requestLogger(ctx, "UpdateUserByUUID")
	var uri request.UUIDParam
	if err := ctx.ShouldBindUri(&uri); err != nil {
		log.Warn("invalid uuid parameter", slog.String("error", err.Error()))
		ctx.JSON(http.StatusBadRequest, response.Error{Error: errInvalidUUID})
		return
	}

	parsedUUID, err := uuid.Parse(uri.UUID)
	if err != nil {
		log.Warn("failed to parse uuid", slog.String("request.uuid_raw", uri.UUID))
		ctx.JSON(http.StatusBadRequest, response.Error{Error: errInvalidUUID})
		return
	}

	var req request.UpdateUser
	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Warn("invalid request body", slog.String("error", err.Error()))
		ctx.JSON(http.StatusBadRequest, response.Error{Error: errInvalidBody})
		return
	}

	log = log.With(
		slog.String("request.user_uuid", parsedUUID.String()),
		slog.Bool("request.username_update", req.Username != nil),
		slog.Bool("request.email_update", req.Email != nil),
		slog.Bool("request.full_name_update", req.FullName != nil),
	)

	updated, err := c.service.UpdateByUUID(parsedUUID, service.UpdateUserInput{
		Username: req.Username,
		Email:    req.Email,
		FullName: req.FullName,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidUserInput):
			log.Warn("invalid user input", slog.String("error", err.Error()))
			ctx.JSON(http.StatusBadRequest, response.Error{Error: err.Error()})
			return
		case errors.Is(err, service.ErrUserNotFound):
			log.Warn("user not found", slog.String("error", err.Error()))
			ctx.JSON(http.StatusNotFound, response.Error{Error: err.Error()})
			return
		case errors.Is(err, service.ErrUserAlreadyExists):
			log.Warn("user already exists", slog.String("error", err.Error()))
			ctx.JSON(http.StatusConflict, response.Error{Error: err.Error()})
			return
		default:
			log.Error("failed to update user by uuid", slog.String("error", err.Error()))
			ctx.JSON(http.StatusInternalServerError, response.Error{Error: err.Error()})
			return
		}
	}

	log.Info("user updated by uuid", slog.Int("user.id", updated.ID))
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
	log := c.requestLogger(ctx, "DeleteUserByUUID")
	var uri request.UUIDParam
	if err := ctx.ShouldBindUri(&uri); err != nil {
		log.Warn("invalid uuid parameter", slog.String("error", err.Error()))
		ctx.JSON(http.StatusBadRequest, response.Error{Error: errInvalidUUID})
		return
	}

	parsedUUID, err := uuid.Parse(uri.UUID)
	if err != nil {
		log.Warn("failed to parse uuid", slog.String("request.uuid_raw", uri.UUID))
		ctx.JSON(http.StatusBadRequest, response.Error{Error: errInvalidUUID})
		return
	}

	log = log.With(slog.String("request.user_uuid", parsedUUID.String()))

	if err := c.service.DeleteByUUID(parsedUUID); err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidUserInput):
			log.Warn("invalid user input", slog.String("error", err.Error()))
			ctx.JSON(http.StatusBadRequest, response.Error{Error: err.Error()})
			return
		case errors.Is(err, service.ErrUserNotFound):
			log.Warn("user not found", slog.String("error", err.Error()))
			ctx.JSON(http.StatusNotFound, response.Error{Error: err.Error()})
			return
		default:
			log.Error("failed to delete user by uuid", slog.String("error", err.Error()))
			ctx.JSON(http.StatusInternalServerError, response.Error{Error: err.Error()})
			return
		}
	}

	log.Info("user deleted by uuid")
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
	log := c.requestLogger(ctx, "UpdateUserByID")
	var uri request.IDParam
	if err := ctx.ShouldBindUri(&uri); err != nil {
		log.Warn("invalid id parameter", slog.String("error", err.Error()))
		ctx.JSON(http.StatusBadRequest, response.Error{Error: errInvalidID})
		return
	}

	var req request.UpdateUser
	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Warn("invalid request body", slog.String("error", err.Error()))
		ctx.JSON(http.StatusBadRequest, response.Error{Error: errInvalidBody})
		return
	}

	log = log.With(
		slog.Int64("request.user_id", uri.ID),
		slog.Bool("request.username_update", req.Username != nil),
		slog.Bool("request.email_update", req.Email != nil),
		slog.Bool("request.full_name_update", req.FullName != nil),
	)

	updated, err := c.service.UpdateByID(uri.ID, service.UpdateUserInput{
		Username: req.Username,
		Email:    req.Email,
		FullName: req.FullName,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidUserInput):
			log.Warn("invalid user input", slog.String("error", err.Error()))
			ctx.JSON(http.StatusBadRequest, response.Error{Error: err.Error()})
			return
		case errors.Is(err, service.ErrUserNotFound):
			log.Warn("user not found", slog.String("error", err.Error()))
			ctx.JSON(http.StatusNotFound, response.Error{Error: err.Error()})
			return
		case errors.Is(err, service.ErrUserAlreadyExists):
			log.Warn("user already exists", slog.String("error", err.Error()))
			ctx.JSON(http.StatusConflict, response.Error{Error: err.Error()})
			return
		default:
			log.Error("failed to update user by id", slog.String("error", err.Error()))
			ctx.JSON(http.StatusInternalServerError, response.Error{Error: err.Error()})
			return
		}
	}

	log.Info("user updated by id", slog.String("user.uuid", updated.UUID))
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
	log := c.requestLogger(ctx, "DeleteUserByID")
	var uri request.IDParam
	if err := ctx.ShouldBindUri(&uri); err != nil {
		log.Warn("invalid id parameter", slog.String("error", err.Error()))
		ctx.JSON(http.StatusBadRequest, response.Error{Error: errInvalidID})
		return
	}

	log = log.With(slog.Int64("request.user_id", uri.ID))

	if err := c.service.DeleteByID(uri.ID); err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidUserInput):
			log.Warn("invalid user input", slog.String("error", err.Error()))
			ctx.JSON(http.StatusBadRequest, response.Error{Error: err.Error()})
			return
		case errors.Is(err, service.ErrUserNotFound):
			log.Warn("user not found", slog.String("error", err.Error()))
			ctx.JSON(http.StatusNotFound, response.Error{Error: err.Error()})
			return
		default:
			log.Error("failed to delete user by id", slog.String("error", err.Error()))
			ctx.JSON(http.StatusInternalServerError, response.Error{Error: err.Error()})
			return
		}
	}

	log.Info("user deleted by id")
	ctx.Status(http.StatusNoContent)
}
