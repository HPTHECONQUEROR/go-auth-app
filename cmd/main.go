package main

import (
	"go-auth-app/config"
	"go-auth-app/db"
	"go-auth-app/internal/delivery"
	"go-auth-app/internal/repository"
	"go-auth-app/internal/routes"
	"go-auth-app/internal/usecase"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load Config
	config.LoadEnv()

	// Initialize config
	cfg := config.LoadConfig()

	// Initialize the DB
	db.ConnectDB(cfg)
	defer db.CloseDB()

	// Run database migrations
	db.RunMigrations()

	// Initialize repositories
	userRepo := repository.NewUserRepository()
	chatRepo := repository.NewChatRepository()

	// Initialize usecases
	authUsecase := usecase.NewAuthUsecase(userRepo)
	chatUsecase := usecase.NewChatUsecase(chatRepo, userRepo)

	// Initialize handlers
	authHandler := delivery.NewAuthHandler(authUsecase)
	chatHandler := delivery.NewChatHandler(chatUsecase)
	wsHandler := delivery.NewWebSocketHandler(chatUsecase)

	// Initialize and configure router
	router := gin.Default()
	
	// Setup middleware
	router.Use(delivery.ErrorHandlerMiddleware())
	
	// Setup routes
	routes.SetupRoutes(router, authHandler, chatHandler, wsHandler)
	
	// Start server
	router.Run(":" + cfg.Port)
}