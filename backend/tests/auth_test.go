package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mradu/task-manager/internal/api"
	"github.com/mradu/task-manager/internal/auth"
)

func TestGenerateAndValidateToken(t *testing.T) {
	userID := 1
	email := "test@example.com"
	role := "user"

	token, err := auth.GenerateToken(userID, email, role)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if token == "" {
		t.Fatalf("Expected token, got empty string")
	}

	claims, err := auth.ValidateToken(token)
	if err != nil {
		t.Fatalf("Expected no error on validation, got %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("Expected UserID %d, got %d", userID, claims.UserID)
	}
	if claims.Email != email {
		t.Errorf("Expected Email %s, got %s", email, claims.Email)
	}
}

func TestAuthMiddleware_MissingToken(t *testing.T) {
	req, _ := http.NewRequest("GET", "/protected", nil)
	rr := httptest.NewRecorder()

	handler := api.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	token, _ := auth.GenerateToken(1, "test@example.com", "user")

	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	handler := api.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := api.GetUserFromContext(r.Context())
		if user == nil {
			t.Errorf("Expected user in context, got nil")
		}
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}
}
