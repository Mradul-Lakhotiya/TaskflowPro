package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/mradu/task-manager/internal/database"
)

type User struct {
	ID           int    `json:"id"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
	Role         string `json:"role"`
}

var ErrUserNotFound = errors.New("user not found")
var ErrEmailExists = errors.New("email already in use")

func CreateUser(ctx context.Context, email, passwordHash, role string) (*User, error) {
	var user User
	err := database.DB.QueryRow(ctx,
		"INSERT INTO users (email, password_hash, role) VALUES ($1, $2, $3) RETURNING id, email, password_hash, role",
		email, passwordHash, role,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role)

	if err != nil {
		// Basic check for unique violation
		if err.Error() == "ERROR: duplicate key value violates unique constraint \"users_email_key\" (SQLSTATE 23505)" {
			return nil, ErrEmailExists
		}
		return nil, err
	}
	return &user, nil
}

func GetUserByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	err := database.DB.QueryRow(ctx,
		"SELECT id, email, password_hash, role FROM users WHERE email = $1",
		email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}
