package db

import (
	"context"
	"fmt"
	"log"
)

func RunMigrations(){
	migrationSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		email VARCHAR(100) UNIQUE NOT NULL,
		password TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`
	_,err := DB.Exec(context.Background(),migrationSQL)
	if err != nil{
		log.Fatal("sorry, Migration failed due to", err)
	}
	fmt.Print("DB migrated successfully!")
}