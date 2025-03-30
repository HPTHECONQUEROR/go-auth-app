package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"go-auth-app/config"

	"github.com/jackc/pgx/v5/pgxpool"

)


var DB *pgxpool.Pool

func ConnectDB(cfg *config.Config){
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
	cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBSSLMode)

	var err error
	DB, err = pgxpool.New(context.Background(),dsn)
	if err!=nil{
		log.Fatal("Unable to connect to DB: ",err)
	}

	ctx,cancel := context.WithTimeout(context.Background(),5*time.Second)
	defer cancel()
	if err := DB.Ping(ctx); err != nil{
		log.Fatal("DB connec test Failed:",err)
	}

	log.Println("Connnected to Postgresql!")
}