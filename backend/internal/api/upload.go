package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mradu/task-manager/internal/repository"
)

// UploadAttachmentHandler handles file uploads for tasks
func UploadAttachmentHandler(w http.ResponseWriter, r *http.Request) {
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

	// Limit upload size to 10MB
	r.ParseMultipartForm(10 << 20)

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Ensure uploads directory exists
	uploadDir := "uploads"
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		http.Error(w, "Server error creating upload directory", http.StatusInternalServerError)
		return
	}

	// Generate safe, unique filename
	filename := fmt.Sprintf("%d_%d_%s", taskID, time.Now().Unix(), header.Filename)
	savePath := filepath.Join(uploadDir, filename)

	dst, err := os.Create(savePath)
	if err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}

	// Generate public URL
	attachmentURL := fmt.Sprintf("/uploads/%s", filename)

	// Update the database
	updatedTask, err := repository.UploadAttachment(r.Context(), taskID, user.UserID, user.Role, header.Filename, attachmentURL)
	if err != nil {
		// Clean up the file if DB update fails
		os.Remove(savePath)
		if err == repository.ErrTaskNotFound {
			http.Error(w, "Task not found or unauthorized", http.StatusNotFound)
			return
		}
		http.Error(w, "Error updating task", http.StatusInternalServerError)
		return
	}

	repository.LogActivity(r.Context(), taskID, user.UserID, "attachment_uploaded", "Attached file: "+header.Filename)

	// If SSE is active, broadcast it
	select {
	case AppHub.broadcast <- SSEEvent{
		Type:   "TASK_UPDATED",
		Task:   updatedTask,
		UserID: updatedTask.UserID,
	}:
	default:
		// Do nothing if SSE is disabled/full
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedTask)
}

// DeleteAttachmentHandler handles removing a file attachment
func DeleteAttachmentHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	attachmentIDStr := chi.URLParam(r, "attachmentId")
	attachmentID, err := strconv.Atoi(attachmentIDStr)
	if err != nil {
		http.Error(w, "Invalid attachment ID", http.StatusBadRequest)
		return
	}

	updatedTask, err := repository.DeleteAttachment(r.Context(), attachmentID, user.UserID, user.Role)
	if err != nil {
		if err == repository.ErrTaskNotFound {
			http.Error(w, "Attachment not found or unauthorized", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to delete attachment", http.StatusInternalServerError)
		return
	}

	// We don't delete the physical file here to keep it simple, 
	// or we could fetch the URL first and os.Remove it.
	
	repository.LogActivity(r.Context(), updatedTask.ID, user.UserID, "attachment_deleted", "Deleted an attachment")

	// If SSE is active, broadcast it
	select {
	case AppHub.broadcast <- SSEEvent{
		Type:   "TASK_UPDATED",
		Task:   updatedTask,
		UserID: updatedTask.UserID,
	}:
	default:
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedTask)
}
