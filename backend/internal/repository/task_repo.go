package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/mradu/task-manager/internal/database"
)

type Task struct {
	ID          int        `json:"id"`
	UserID        int        `json:"user_id"`
	Title         string     `json:"title"`
	Description   string     `json:"description"`
	Status        string     `json:"status"`
	Priority      string     `json:"priority"`
	AttachmentURL *string    `json:"attachment_url"`
	DueDate       *time.Time `json:"due_date"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type TaskFilter struct {
	UserID       int
	Role         string // "admin" can see all tasks if we want, but for now we filter by UserID if not admin
	FilterUserID int    // specific user to filter by for admins
	Status       string
	Search       string
	SortBy       string // due_date, priority, created_at
	SortDesc     bool
	Page         int
	Limit        int
}

var ErrTaskNotFound = errors.New("task not found or unauthorized")

func CreateTask(ctx context.Context, t *Task) (*Task, error) {
	err := database.DB.QueryRow(ctx,
		"INSERT INTO tasks (user_id, title, description, status, priority, attachment_url, due_date) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, created_at, updated_at",
		t.UserID, t.Title, t.Description, t.Status, t.Priority, t.AttachmentURL, t.DueDate,
	).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)

	if err != nil {
		return nil, err
	}
	return t, nil
}

func GetTaskByID(ctx context.Context, id, userID int, role string) (*Task, error) {
	var task Task
	err := database.DB.QueryRow(ctx,
		`SELECT id, user_id, title, description, status, priority, attachment_url, due_date, created_at, updated_at 
		 FROM tasks WHERE id = $1 AND ($2 = 'admin' OR user_id = $3)`,
		id, role, userID,
	).Scan(
		&task.ID, &task.UserID, &task.Title, &task.Description,
		&task.Status, &task.Priority, &task.AttachmentURL, &task.DueDate,
		&task.CreatedAt, &task.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}
	return &task, nil
}

func UpdateTask(ctx context.Context, id, userID int, role string, updates map[string]interface{}) (*Task, error) {
	if len(updates) == 0 {
		return GetTaskByID(ctx, id, userID, role)
	}

	setParts := []string{}
	args := []interface{}{}
	argID := 1

	for k, v := range updates {
		setParts = append(setParts, fmt.Sprintf("%s = $%d", k, argID))
		args = append(args, v)
		argID++
	}

	// Always update updated_at
	setParts = append(setParts, fmt.Sprintf("updated_at = CURRENT_TIMESTAMP"))

	query := fmt.Sprintf(`UPDATE tasks SET %s WHERE id = $%d`, strings.Join(setParts, ", "), argID)
	args = append(args, id)
	argID++

	if role != "admin" {
		query += fmt.Sprintf(` AND user_id = $%d`, argID)
		args = append(args, userID)
	}

	query += ` RETURNING id, user_id, title, description, status, priority, attachment_url, due_date, created_at, updated_at`

	var t Task
	err := database.DB.QueryRow(ctx, query, args...).Scan(
		&t.ID, &t.UserID, &t.Title, &t.Description, &t.Status, &t.Priority, &t.AttachmentURL, &t.DueDate, &t.CreatedAt, &t.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}

	return &t, nil
}

func DeleteTask(ctx context.Context, id, userID int, role string) error {
	query := `DELETE FROM tasks WHERE id = $1`
	args := []interface{}{id}

	if role != "admin" {
		query += ` AND user_id = $2`
		args = append(args, userID)
	}

	cmdTag, err := database.DB.Exec(ctx, query, args...)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return ErrTaskNotFound
	}
	return nil
}

func ListTasks(ctx context.Context, filter TaskFilter) ([]Task, int, error) {
	whereParts := []string{"1=1"}
	args := []interface{}{}
	argID := 1

	if filter.Role != "admin" {
		whereParts = append(whereParts, fmt.Sprintf("user_id = $%d", argID))
		args = append(args, filter.UserID)
		argID++
	} else if filter.FilterUserID > 0 {
		whereParts = append(whereParts, fmt.Sprintf("user_id = $%d", argID))
		args = append(args, filter.FilterUserID)
		argID++
	}

	if filter.Status != "" {
		whereParts = append(whereParts, fmt.Sprintf("status = $%d", argID))
		args = append(args, filter.Status)
		argID++
	}

	if filter.Search != "" {
		whereParts = append(whereParts, fmt.Sprintf("title ILIKE $%d", argID))
		args = append(args, "%"+filter.Search+"%")
		argID++
	}

	whereClause := strings.Join(whereParts, " AND ")

	// Count total rows for pagination
	var total int
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM tasks WHERE %s`, whereClause)
	if err := database.DB.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Sorting
	orderClause := "created_at DESC"
	if filter.SortBy != "" {
		allowedSorts := map[string]string{
			"due_date":   "due_date",
			"priority":   "priority",
			"created_at": "created_at",
		}
		if col, ok := allowedSorts[filter.SortBy]; ok {
			dir := "ASC"
			if filter.SortDesc {
				dir = "DESC"
			}
			orderClause = fmt.Sprintf("%s %s", col, dir)
			
			// Handle priority custom sorting (high > medium > low)
			if col == "priority" {
				orderClause = fmt.Sprintf("CASE priority WHEN 'high' THEN 1 WHEN 'medium' THEN 2 WHEN 'low' THEN 3 ELSE 4 END %s", dir)
			}
			// Handle nulls last for due date
			if col == "due_date" {
				orderClause += " NULLS LAST"
			}
		}
	}

	// Pagination
	if filter.Limit <= 0 {
		filter.Limit = 10
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}
	offset := (filter.Page - 1) * filter.Limit

	query := fmt.Sprintf(`
		SELECT id, user_id, title, description, status, priority, attachment_url, due_date, created_at, updated_at 
		FROM tasks 
		WHERE %s 
		ORDER BY %s 
		LIMIT $%d OFFSET $%d`, 
		whereClause, orderClause, argID, argID+1)
	
	args = append(args, filter.Limit, offset)

	rows, err := database.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(
			&t.ID, &t.UserID, &t.Title, &t.Description, 
			&t.Status, &t.Priority, &t.AttachmentURL, &t.DueDate, 
			&t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		tasks = append(tasks, t)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}
