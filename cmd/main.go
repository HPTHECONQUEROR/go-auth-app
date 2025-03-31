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
	//Load COnfig
	config.LoadEnv()

	//Initialize config
	cfg := config.LoadConfig()

	//Initialise the DB
	db.ConnectDB(cfg)
	defer db.CloseDB()

	db.RunMigrations()

	userRepo := repository.NewUserRepository()
	authUsecase := usecase.NewAuthUsecase(userRepo)  
	authHandler := delivery.NewAuthHandler(authUsecase)  

	router := gin.Default()
	routes.SetupRoutes(router,authHandler)
	router.Run(":8000")
}
