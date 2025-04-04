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
	// "os/exec"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load Config
	config.LoadEnv()

	// nats_file := "nats-server.exe"
	// exec.Command(nats_file)

	// Initialize config
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



	// Initialize NATS service
	natsService := service.NewNATSService(natsClient)

	// Generate mock data for testing NATS
	// natsService.MockChatData()

	// Initialize repositories
	userRepo := repository.NewUserRepository()
	chatRepo := repository.NewChatRepository()

	// Initialize usecases
	authUsecase := usecase.NewAuthUsecase(userRepo)
	chatUsecase := usecase.NewChatUsecase(chatRepo, userRepo, natsService)

	// Initialize handlers
	authHandler := delivery.NewAuthHandler(authUsecase)
	chatHandler := delivery.NewChatHandler(chatUsecase)
	wsHandler := delivery.NewWebSocketHandler(chatUsecase, natsService)

	// Initialize and configure router
	router := gin.Default()

	// Setup middleware
	router.Use(delivery.ErrorHandlerMiddleware())

	// Setup routes

	routes.SetupRoutes(router, authHandler, chatHandler, wsHandler)

	// Log server startup
	log.Printf("Starting server on port %s", cfg.Port)

	// Start server
	router.Run(":" + cfg.Port)
}
