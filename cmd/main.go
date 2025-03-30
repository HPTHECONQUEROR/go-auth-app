package main

import (
	"fmt"
	"go-auth-app/config"
	"go-auth-app/db"
)

func main() {
	cfg := config.LoadConfig()
	db.ConnectDB(cfg)
	fmt.Println("Server Running on the port:", cfg.Port)
	// fmt.Println("DB Host:", cfg.DBHost)
	// fmt.Println("JWT Secret Key:", cfg.JWTSecret)
	db.RunMigrations()
	fmt.Println("Migrations are done")
}
