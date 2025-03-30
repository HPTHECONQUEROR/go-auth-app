package main

import (
	"fmt"
	"go-auth-app/config"
	"go-auth-app/db"
	"go-auth-app/internal/delivery"
	"go-auth-app/internal/repository"
	"go-auth-app/internal/usecase"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()
	db.ConnectDB(cfg)
	fmt.Println("Server Running on the port:", cfg.Port)
	// fmt.Println("DB Host:", cfg.DBHost)
	// fmt.Println("JWT Secret Key:", cfg.JWTSecret)
	db.RunMigrations()
	// fmt.Println("Migrations are done")

	userRepo := repository.NewUserRepository()
	authUsecase := usecase.NewAuthUsecase(*userRepo)  
	authHandler := delivery.NewAuthHandler(authUsecase)  

	router := gin.Default()

	router.POST("/signup",authHandler.SignupHandler)
	router.POST("/login",authHandler.LoginHandler)
	router.GET("/protected", delivery.AuthMiddleware(), func(c *gin.Context) {
		userID := c.GetInt("user_id")
		email := c.GetString("email")
		c.JSON(http.StatusOK, gin.H{
			"message":"Protected Data",
			"user_id":userID,
			"email":email,
		})
	})

	router.Run(":8000")
}
