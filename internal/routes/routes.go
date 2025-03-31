package routes

import (
	"go-auth-app/internal/delivery"
	"github.com/gin-gonic/gin"
)

// SetupRoutes defines API routes
func SetupRoutes(router *gin.Engine, authHandler *delivery.AuthHandler) {
	// Public Routes
	router.POST("/signup", authHandler.SignupHandler)
	router.POST("/login", authHandler.LoginHandler)

	// Protected Routes
	router.GET("/protected", delivery.AuthMiddleware(), authHandler.ProtectedHandler)
}