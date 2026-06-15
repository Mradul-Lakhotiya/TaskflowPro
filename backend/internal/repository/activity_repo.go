package repository

import (
	"context"
	"time"

	"github.com/mradu/task-manager/internal/database"
)

type TaskActivity struct {
	ID        int       `json:"id"`
	TaskID    int       `json:"task_id"`
	UserID    *int      `json:"user_id"`
	UserEmail string    `json:"user_email,omitempty"`
	Action    string    `json:"action"`
	Details   *string   `json:"details"`
	CreatedAt time.Time `json:"created_at"`
}

func LogActivity(ctx context.Context, taskID int, userID int, action string, details string) error {
	var detailsPtr *string
	if details != "" {
		detailsPtr = &details
	}

	query := `
		INSERT INTO task_activities (task_id, user_id, action, details)
		VALUES ($1, $2, $3, $4)
	`
	_, err := database.DB.Exec(ctx, query, taskID, userID, action, detailsPtr)
	return err
}

func GetTaskActivity(ctx context.Context, taskID int) ([]TaskActivity, error) {
	query := `
		SELECT a.id, a.task_id, a.user_id, u.email as user_email, a.action, a.details, a.created_at
		FROM task_activities a
		LEFT JOIN users u ON a.user_id = u.id
		WHERE a.task_id = $1
		ORDER BY a.created_at DESC
	`

	rows, err := database.DB.Query(ctx, query, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []TaskActivity
	for rows.Next() {
		var a TaskActivity
		var email *string
		if err := rows.Scan(&a.ID, &a.TaskID, &a.UserID, &email, &a.Action, &a.Details, &a.CreatedAt); err != nil {
			return nil, err
		}
		if email != nil {
			a.UserEmail = *email
		}
		activities = append(activities, a)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return activities, nil
}
