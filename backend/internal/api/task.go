package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mradu/task-manager/internal/repository"
)

type CreateTaskRequest struct {
	Title       string     `json:"title" validate:"required"`
	Description string     `json:"description"`
	Status      string     `json:"status" validate:"omitempty,oneof=pending in_progress completed"`
	Priority    string     `json:"priority" validate:"omitempty,oneof=low medium high"`
	DueDate     *time.Time `json:"due_date"`
}

type UpdateTaskRequest struct {
	Title       *string    `json:"title" validate:"omitempty,min=1"`
	Description *string    `json:"description"`
	Status      *string    `json:"status" validate:"omitempty,oneof=pending in_progress completed"`
	Priority    *string    `json:"priority" validate:"omitempty,oneof=low medium high"`
	DueDate     *time.Time `json:"due_date"`
}

func CreateTaskHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if err := validate.Struct(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	status := "pending"
	if req.Status != "" {
		status = req.Status
	}
	priority := "medium"
	if req.Priority != "" {
		priority = req.Priority
	}

	task := &repository.Task{
		UserID:      user.UserID,
		Title:       req.Title,
		Description: req.Description,
		Status:      status,
		Priority:    priority,
		DueDate:     req.DueDate,
	}

	createdTask, err := repository.CreateTask(r.Context(), task)
	if err != nil {
		http.Error(w, "Failed to create task", http.StatusInternalServerError)
		return
	}

	// Log activity
	repository.LogActivity(r.Context(), createdTask.ID, user.UserID, "created", "Task was created")

	AppHub.broadcast <- SSEEvent{
		Type:   "TASK_CREATED",
		Task:   createdTask,
		UserID: createdTask.UserID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdTask)
}

func GetTasksHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	q := r.URL.Query()
	
	page, _ := strconv.Atoi(q.Get("page"))
	if page == 0 {
		page = 1
	}
	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit == 0 {
		limit = 10
	}

	sortDesc := false
	if q.Get("sort_desc") == "true" {
		sortDesc = true
	}

	filterUserID, _ := strconv.Atoi(q.Get("user_id"))

	filter := repository.TaskFilter{
		UserID:       user.UserID,
		Role:         user.Role,
		FilterUserID: filterUserID,
		Status:       q.Get("status"),
		Search:       q.Get("search"),
		SortBy:       q.Get("sort_by"),
		SortDesc:     sortDesc,
		Page:         page,
		Limit:        limit,
	}

	tasks, total, err := repository.ListTasks(r.Context(), filter)
	if err != nil {
		http.Error(w, "Failed to fetch tasks", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"data":  tasks,
		"total": total,
		"page":  page,
		"limit": limit,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func GetTaskByIDHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	taskIDStr := chi.URLParam(r, "id")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	task, err := repository.GetTaskByID(r.Context(), taskID, user.UserID, user.Role)
	if err != nil {
		if err == repository.ErrTaskNotFound {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to fetch task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func UpdateTaskHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	taskIDStr := chi.URLParam(r, "id")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	oldTask, err := repository.GetTaskByID(r.Context(), taskID, user.UserID, user.Role)
	if err != nil {
		if err == repository.ErrTaskNotFound {
			http.Error(w, "Task not found or unauthorized", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to fetch task", http.StatusInternalServerError)
		return
	}

	var req UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if err := validate.Struct(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	updates := make(map[string]interface{})
	var details []string

	if req.Title != nil && *req.Title != oldTask.Title {
		updates["title"] = *req.Title
		details = append(details, "Title updated")
	}
	if req.Description != nil && *req.Description != oldTask.Description {
		updates["description"] = *req.Description
		details = append(details, "Description updated")
	}
	if req.Status != nil && *req.Status != oldTask.Status {
		updates["status"] = *req.Status
		details = append(details, "Status changed to "+*req.Status)
	}
	if req.Priority != nil && *req.Priority != oldTask.Priority {
		updates["priority"] = *req.Priority
		details = append(details, "Priority changed to "+*req.Priority)
	}
	if req.DueDate != nil {
		updates["due_date"] = req.DueDate
		details = append(details, "Due date updated")
	}

	updatedTask, err := repository.UpdateTask(r.Context(), taskID, user.UserID, user.Role, updates)
	if err != nil {
		if err == repository.ErrTaskNotFound {
			http.Error(w, "Task not found or unauthorized", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to update task", http.StatusInternalServerError)
		return
	}

	if len(details) > 0 {
		detailStr := ""
		for i, d := range details {
			if i > 0 {
				detailStr += ", "
			}
			detailStr += d
		}
		repository.LogActivity(r.Context(), taskID, user.UserID, "updated", detailStr)
	}

	AppHub.broadcast <- SSEEvent{
		Type:   "TASK_UPDATED",
		Task:   updatedTask,
		UserID: updatedTask.UserID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedTask)
}

func DeleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	taskIDStr := chi.URLParam(r, "id")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	// Fetch the task first to determine its owner before deleting
	task, fetchErr := repository.GetTaskByID(r.Context(), taskID, user.UserID, user.Role)
	if fetchErr != nil {
		if fetchErr == repository.ErrTaskNotFound {
			http.Error(w, "Task not found or unauthorized", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to delete task", http.StatusInternalServerError)
		return
	}

	err = repository.DeleteTask(r.Context(), taskID, user.UserID, user.Role)
	if err != nil {
		http.Error(w, "Failed to delete task", http.StatusInternalServerError)
		return
	}

	AppHub.broadcast <- SSEEvent{
		Type:   "TASK_DELETED",
		TaskID: taskID,
		UserID: task.UserID,
	}

	w.WriteHeader(http.StatusNoContent)
}

func GetTaskActivityHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	taskIDStr := chi.URLParam(r, "id")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	// Verify task exists and user has access
	_, err = repository.GetTaskByID(r.Context(), taskID, user.UserID, user.Role)
	if err != nil {
		if err == repository.ErrTaskNotFound {
			http.Error(w, "Task not found or unauthorized", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to fetch task", http.StatusInternalServerError)
		return
	}

	activities, err := repository.GetTaskActivity(r.Context(), taskID)
	if err != nil {
		http.Error(w, "Failed to fetch activities", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(activities)
}
