package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
)

type Config struct {
	Port          string
	DBHost        string
	DBPort        string
	DBUser        string
	DBPassword    string
	DBName        string
	DBSSLMode     string
	JWTSecret     string
	JWTExpiration string
	NatsURL       string
	NatsReconnect bool
}

func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: No .env file found!")
	}
}

func Getenv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: No .env file found, using environment variables")
	}

	return &Config{
		Port:          Getenv("PORT", "8000"),
		DBHost:        Getenv("DB_HOST", "localhost"),
		DBPort:        Getenv("DB_PORT", "5432"),
		DBUser:        Getenv("DB_USER", "postgres"),
		DBPassword:    Getenv("DB_PASSWORD", "8056"),
		DBName:        Getenv("DB_NAME", "go_auth_db"),
		DBSSLMode:     Getenv("DB_SSLMODE", "disable"),
		JWTSecret:     Getenv("JWT_SECRET", "default_jwt_secret_key"),
		JWTExpiration: Getenv("JWT_EXPIRATION_HOURS", "24"),
		NatsURL:       Getenv("NATS_URL", "nats://localhost:4222"),
		NatsReconnect: GetenvBool("NATS_RECONNECT", true),
	}
}

func GetenvBool(key string, fallback bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			return fallback
		}
		return boolValue
	}
	return fallback
}
