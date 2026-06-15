package api

import (
	"encoding/json"
	"net/http"

	"github.com/mradu/task-manager/internal/repository"
)

// GetUsersHandler returns a list of all users. Accessible only by admins.
func GetUsersHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil || user.Role != "admin" {
		http.Error(w, "Unauthorized or forbidden", http.StatusForbidden)
		return
	}

	users, err := repository.ListUsers(r.Context())
	if err != nil {
		http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}
