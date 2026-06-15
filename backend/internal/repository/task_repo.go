package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/mradu/task-manager/internal/database"
)

type TaskAttachment struct {
	ID        int       `json:"id"`
	TaskID    int       `json:"task_id"`
	FileName  string    `json:"file_name"`
	FileURL   string    `json:"file_url"`
	CreatedAt time.Time `json:"created_at"`
}

type Task struct {
	ID          int              `json:"id"`
	UserID      int              `json:"user_id"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Status      string           `json:"status"`
	Priority    string           `json:"priority"`
	Attachments []TaskAttachment `json:"attachments"`
	DueDate     *time.Time       `json:"due_date"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

type TaskFilter struct {
	UserID       int
	Role         string
	FilterUserID int
	Status       string
	Search       string
	SortBy       string
	SortDesc     bool
	Page         int
	Limit        int
}

var ErrTaskNotFound = errors.New("task not found or unauthorized")

func CreateTask(ctx context.Context, t *Task) (*Task, error) {
	err := database.DB.QueryRow(ctx,
		"INSERT INTO tasks (user_id, title, description, status, priority, due_date) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at, updated_at",
		t.UserID, t.Title, t.Description, t.Status, t.Priority, t.DueDate,
	).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)

	if err != nil {
		return nil, err
	}
	t.Attachments = []TaskAttachment{} // New tasks have no attachments
	return t, nil
}

func GetTaskByID(ctx context.Context, id, userID int, role string) (*Task, error) {
	query := `
		SELECT t.id, t.user_id, t.title, t.description, t.status, t.priority, t.due_date, t.created_at, t.updated_at,
		       COALESCE(
		           json_agg(
		               json_build_object('id', a.id, 'task_id', a.task_id, 'file_name', a.file_name, 'file_url', a.file_url, 'created_at', a.created_at)
		           ) FILTER (WHERE a.id IS NOT NULL), '[]'
		       ) as attachments
		FROM tasks t
		LEFT JOIN task_attachments a ON t.id = a.task_id
		WHERE t.id = $1 AND ($2 = 'admin' OR t.user_id = $3)
		GROUP BY t.id
	`
	var task Task
	var attachmentsJSON []byte

	err := database.DB.QueryRow(ctx, query, id, role, userID).Scan(
		&task.ID, &task.UserID, &task.Title, &task.Description,
		&task.Status, &task.Priority, &task.DueDate,
		&task.CreatedAt, &task.UpdatedAt, &attachmentsJSON,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}

	if err := json.Unmarshal(attachmentsJSON, &task.Attachments); err != nil {
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

	setParts = append(setParts, fmt.Sprintf("updated_at = CURRENT_TIMESTAMP"))

	query := fmt.Sprintf(`UPDATE tasks SET %s WHERE id = $%d`, strings.Join(setParts, ", "), argID)
	args = append(args, id)
	argID++

	if role != "admin" {
		query += fmt.Sprintf(` AND user_id = $%d`, argID)
		args = append(args, userID)
	}

	cmdTag, err := database.DB.Exec(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	if cmdTag.RowsAffected() == 0 {
		return nil, ErrTaskNotFound
	}

	return GetTaskByID(ctx, id, userID, role)
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
		whereParts = append(whereParts, fmt.Sprintf("t.user_id = $%d", argID))
		args = append(args, filter.UserID)
		argID++
	} else if filter.FilterUserID > 0 {
		whereParts = append(whereParts, fmt.Sprintf("t.user_id = $%d", argID))
		args = append(args, filter.FilterUserID)
		argID++
	}

	if filter.Status != "" {
		whereParts = append(whereParts, fmt.Sprintf("t.status = $%d", argID))
		args = append(args, filter.Status)
		argID++
	}

	if filter.Search != "" {
		whereParts = append(whereParts, fmt.Sprintf("t.title ILIKE $%d", argID))
		args = append(args, "%"+filter.Search+"%")
		argID++
	}

	whereClause := strings.Join(whereParts, " AND ")

	var total int
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM tasks t WHERE %s`, whereClause)
	if err := database.DB.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	orderClause := "t.created_at DESC"
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
			orderClause = fmt.Sprintf("t.%s %s", col, dir)
			
			if col == "priority" {
				orderClause = fmt.Sprintf("CASE t.priority WHEN 'high' THEN 1 WHEN 'medium' THEN 2 WHEN 'low' THEN 3 ELSE 4 END %s", dir)
			}
			if col == "due_date" {
				orderClause += " NULLS LAST"
			}
		}
	}

	if filter.Limit <= 0 {
		filter.Limit = 10
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}
	offset := (filter.Page - 1) * filter.Limit

	query := fmt.Sprintf(`
		SELECT t.id, t.user_id, t.title, t.description, t.status, t.priority, t.due_date, t.created_at, t.updated_at,
		       COALESCE(
		           json_agg(
		               json_build_object('id', a.id, 'task_id', a.task_id, 'file_name', a.file_name, 'file_url', a.file_url, 'created_at', a.created_at)
		           ) FILTER (WHERE a.id IS NOT NULL), '[]'
		       ) as attachments
		FROM tasks t
		LEFT JOIN task_attachments a ON t.id = a.task_id
		WHERE %s 
		GROUP BY t.id
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
		var attachmentsJSON []byte
		if err := rows.Scan(
			&t.ID, &t.UserID, &t.Title, &t.Description, 
			&t.Status, &t.Priority, &t.DueDate, 
			&t.CreatedAt, &t.UpdatedAt, &attachmentsJSON,
		); err != nil {
			return nil, 0, err
		}
		if err := json.Unmarshal(attachmentsJSON, &t.Attachments); err != nil {
			return nil, 0, err
		}
		tasks = append(tasks, t)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

// UploadAttachment inserts a new attachment into the task_attachments table
func UploadAttachment(ctx context.Context, taskID int, userID int, role string, fileName string, fileURL string) (*Task, error) {
	// First ensure the task exists and user has access
	_, err := GetTaskByID(ctx, taskID, userID, role)
	if err != nil {
		return nil, err
	}

	// Insert into task_attachments
	_, err = database.DB.Exec(ctx, `
		INSERT INTO task_attachments (task_id, file_name, file_url) 
		VALUES ($1, $2, $3)`,
		taskID, fileName, fileURL,
	)
	if err != nil {
		return nil, err
	}

	// Also touch the updated_at on tasks
	_, _ = database.DB.Exec(ctx, `UPDATE tasks SET updated_at = CURRENT_TIMESTAMP WHERE id = $1`, taskID)

	// Return updated task
	return GetTaskByID(ctx, taskID, userID, role)
}

func DeleteAttachment(ctx context.Context, attachmentID int, userID int, role string) (*Task, error) {
	// First fetch the attachment to ensure it exists and get its taskID
	var taskID int
	err := database.DB.QueryRow(ctx, "SELECT task_id FROM task_attachments WHERE id = $1", attachmentID).Scan(&taskID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}

	// Check if user has access to this task
	_, err = GetTaskByID(ctx, taskID, userID, role)
	if err != nil {
		return nil, err
	}

	// Delete it
	_, err = database.DB.Exec(ctx, "DELETE FROM task_attachments WHERE id = $1", attachmentID)
	if err != nil {
		return nil, err
	}

	// Return updated task
	return GetTaskByID(ctx, taskID, userID, role)
}
