package main

import (
	"go-auth-app/config"
	"go-auth-app/db"
	"go-auth-app/internal/delivery"
	"go-auth-app/internal/repository"
	"go-auth-app/internal/routes"
	"go-auth-app/internal/service"
	"go-auth-app/internal/usecase"
	"go-auth-app/pkg"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load Config
	config.LoadEnv()
	cfg := config.LoadConfig()

	// Initialize the DB
	db.ConnectDB(cfg)
	defer db.CloseDB()

	// Run database migrations
	db.RunMigrations()

	// Initialize NATS client
	natsClient, err := pkg.NewNatsClient(cfg.NatsURL, cfg.NatsReconnect)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer natsClient.Close()

	// Initialize repositories
	userRepo := repository.NewUserRepository()
	chatRepo := repository.NewChatRepository()

	// Initialize NATS-related components
	natsService := service.NewNATSService(natsClient)
	natsUsecase := usecase.NewNATSUsecase(natsClient)

	// Initialize usecases
	authUsecase := usecase.NewAuthUsecase(userRepo)
	chatUsecase := usecase.NewChatUsecase(chatRepo, userRepo, natsService)

	// Initialize handlers
	authHandler := delivery.NewAuthHandler(authUsecase)
	chatHandler := delivery.NewChatHandler(chatUsecase)
	wsHandler := delivery.NewWebSocketHandler(chatUsecase, natsService)
	natsHandler := delivery.NewNATSHandler(natsUsecase)

	// Initialize router
	router := gin.Default()
	router.Use(delivery.ErrorHandlerMiddleware())

	// Setup routes
	routes.SetupRoutes(router, authHandler, chatHandler, wsHandler, natsHandler)

	// Start server
	log.Printf("Starting server on port %s", cfg.Port)
	router.Run(":" + cfg.Port)
}
