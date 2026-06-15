package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all the environment-dependent variables for the application.
type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	FrontendURL string
}

// AppConfig is the global configuration instance
var AppConfig *Config

// Load initializes the application configuration from environment variables.
func Load() {
	// Attempt to load .env file if it exists, but don't crash if it doesn't
	// (in production, variables should be injected directly by the host).
	if err := godotenv.Load("../.env"); err != nil {
		if err := godotenv.Load(".env"); err != nil {
			log.Println("No .env file found, relying on pure environment variables")
		}
	}

	AppConfig = &Config{
		Port:        getEnvOrDefault("PORT", "8080"),
		DatabaseURL: buildDatabaseURL(),
		JWTSecret:   getEnvOrDefault("JWT_SECRET", "super_secret_jwt_key_change_in_production"),
		FrontendURL: getEnvOrDefault("FRONTEND_URL", "http://localhost:3000"),
	}
}

// getEnvOrDefault returns the value of an environment variable or a fallback string.
func getEnvOrDefault(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return fallback
}

// buildDatabaseURL constructs the PostgreSQL connection string.
// It checks for a direct DATABASE_URL first, then falls back to composing it from parts.
func buildDatabaseURL() string {
	if url := os.Getenv("DATABASE_URL"); url != "" {
		return url
	}

	host := getEnvOrDefault("DB_HOST", "localhost")
	port := getEnvOrDefault("DB_PORT", "5432")
	user := getEnvOrDefault("DB_USER", "postgres")
	password := getEnvOrDefault("DB_PASSWORD", "postgres")
	dbname := getEnvOrDefault("DB_NAME", "task_manager")

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbname)
}
