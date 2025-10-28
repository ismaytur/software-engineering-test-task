package controller

import (
	"errors"
	"net/http"
	"strconv"

	"cruder/internal/controller/response"
	"cruder/internal/service"

	"github.com/gin-gonic/gin"
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
// @Failure      404  {object}  response.Error  "user not found"
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
// @Failure      400  {object}  response.Error  "invalid id"
// @Failure      404  {object}  response.Error  "user not found"
// @Failure      500  {object}  response.Error
// @Router       /api/v1/users/id/{id} [get]
func (c *UserController) GetUserByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, response.Error{Error: "invalid id"})
		return
	}

	user, err := c.service.GetByID(id)
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
