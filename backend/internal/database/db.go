package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

// Connect establishes a connection pool to the PostgreSQL database
func Connect(databaseURL string) error {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return fmt.Errorf("error parsing database config: %w", err)
	}

	// Set connection pool limits
	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	// Connect to the database
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return fmt.Errorf("error connecting to database: %w", err)
	}

	// Ping to ensure connection is valid
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("error pinging database: %w", err)
	}

	log.Println("Successfully connected to the database")
	DB = pool
	return nil
}

// Close gracefully closes the connection pool
func Close() {
	if DB != nil {
		DB.Close()
		log.Println("Database connection closed")
	}
}
