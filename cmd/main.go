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
	
	// Setup routes
	routes.SetupRoutes(router, authHandler, chatHandler, wsHandler)
	
	// Start server
	router.Run(":" + cfg.Port)
}



Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InVzcjFAZW1haWwuY29tIiwiZXhwIjoxNzQzNTcwNTI5LCJ1c2VyX2lkIjo2fQ.yCd2Gh6r3kXlZGePLxfNoU7y6tXtCul9zorCjhPpj6s
Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InVzcjJAZW1haWwuY29tIiwiZXhwIjoxNzQzNTcwNTMyLCJ1c2VyX2lkIjo3fQ.Xdl2JwmGcdLuDiAVAYk8zTSr5iJCfCJtQsSDrSQiN00