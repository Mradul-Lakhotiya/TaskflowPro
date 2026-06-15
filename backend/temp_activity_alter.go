package main

import (
	"context"
	"log"

	"github.com/mradu/task-manager/internal/config"
	"github.com/mradu/task-manager/internal/database"
)

func main() {
	config.Load()

	if err := database.Connect(config.AppConfig.DatabaseURL); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	query := `
	CREATE TABLE IF NOT EXISTS task_activities (
		id SERIAL PRIMARY KEY,
		task_id INTEGER NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
		user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
		action VARCHAR(50) NOT NULL,
		details TEXT,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_task_activities_task_id ON task_activities(task_id);
	`

	_, err := database.DB.Exec(context.Background(), query)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	log.Println("task_activities table created successfully.")
}
