package models

import (
	"database/sql"
	"errors"
	"forum/utils"
	"time"
)

type User struct {
	ID        int64
	Username  string
	Email     string
	Password  string
	CreatedAt time.Time
}

// CreateUser creates a new user in the database
func CreateUser(db *sql.DB, username, email, password string) (int64, error) {
	// Check if username already exists
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", username).Scan(&count)
	if err != nil {
		return 0, err
	}
	if count > 0 {
		return 0, errors.New("username already taken")
	}

	// Check if email already exists
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", email).Scan(&count)
	if err != nil {
		return 0, err
	}
	if count > 0 {
		return 0, errors.New("email already registered")
	}

	// Hash the password
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return 0, err
	}

	// Insert the new user
	result, err := db.Exec(
		"INSERT INTO users (username, email, password) VALUES (?, ?, ?)",
		username, email, hashedPassword,
	)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// GetUserByID retrieves a user by ID
func GetUserByID(db *sql.DB, id int64) (*User, error) {
	var user User
	err := db.QueryRow(
		"SELECT id, username, email, password, created_at FROM users WHERE id = ?",
		id,
	).Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// AuthenticateUser authenticates a user with email and password
func AuthenticateUser(db *sql.DB, email, password string) (*User, error) {
	var user User
	err := db.QueryRow(
		"SELECT id, username, email, password, created_at FROM users WHERE email = ?",
		email,
	).Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("invalid email or password")
		}
		return nil, err
	}

	// Verify password
	if !utils.CheckPasswordHash(password, user.Password) {
		return nil, errors.New("invalid email or password")
	}

	return &user, nil
}
