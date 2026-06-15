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

	ctx := context.Background()
	tx, err := database.DB.Begin(ctx)
	if err != nil {
		log.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback(ctx)

	// 1. Create the new task_attachments table
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS task_attachments (
		id SERIAL PRIMARY KEY,
		task_id INTEGER NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
		file_name VARCHAR(255) NOT NULL,
		file_url TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_task_attachments_task_id ON task_attachments(task_id);
	`
	if _, err := tx.Exec(ctx, createTableQuery); err != nil {
		log.Fatalf("Failed to create task_attachments table: %v", err)
	}

	// 2. Migrate existing data (assuming attachment_url holds the URL, and we can extract filename or just use "attachment")
	migrateQuery := `
	INSERT INTO task_attachments (task_id, file_name, file_url)
	SELECT id, 'Attachment', attachment_url
	FROM tasks
	WHERE attachment_url IS NOT NULL;
	`
	if _, err := tx.Exec(ctx, migrateQuery); err != nil {
		log.Fatalf("Failed to migrate data: %v", err)
	}

	// 3. Drop the old column
	dropColumnQuery := `
	ALTER TABLE tasks DROP COLUMN IF EXISTS attachment_url;
	`
	if _, err := tx.Exec(ctx, dropColumnQuery); err != nil {
		log.Fatalf("Failed to drop old column: %v", err)
	}

	// Commit
	if err := tx.Commit(ctx); err != nil {
		log.Fatalf("Failed to commit transaction: %v", err)
	}

	log.Println("Successfully migrated to task_attachments table!")
}
