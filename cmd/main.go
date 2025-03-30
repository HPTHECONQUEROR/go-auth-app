package main

import (
	"fmt"
	"go-auth-app/config"
)

func main() {
	cfg := config.LoadConfig()
	fmt.Println("Server Running on the port:", cfg.Port)
	fmt.Println("DB Host:", cfg.DBHost)
	// fmt.Println("JWT Secret Key:", cfg.JWTSecret)
}
