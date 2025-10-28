package handler

import (
	"cruder/internal/controller"

	"github.com/gin-gonic/gin"
)

func New(router *gin.Engine, userController *controller.UserController) *gin.Engine {
	v1 := router.Group("/api/v1")
	{
		userGroup := v1.Group("/users")
		{
			userGroup.GET("/", userController.GetAllUsers)
			userGroup.GET("/username/:username", userController.GetUserByUsername)
			userGroup.GET("/id/:id", userController.GetUserByID)
			userGroup.GET("/uuid/:uuid", userController.GetUserByUUID)
			userGroup.POST("/", userController.CreateUser)
			userGroup.PATCH("/uuid/:uuid", userController.UpdateUserByUUID)
			userGroup.PATCH("/id/:id", userController.UpdateUserByID)
			userGroup.DELETE("/uuid/:uuid", userController.DeleteUserByUUID)
			userGroup.DELETE("/id/:id", userController.DeleteUserByID)
		}
	}
	return router
}
