package utils

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

// CreateSession creates a new session for a user
// and invalidates any existing sessions for this user
func CreateSession(db *sql.DB, userID int64) (string, error) {
	// Generate a unique session ID
	sessionID := uuid.New().String()

	// Set expiration time (24 hours from now for better security)
	expiresAt := time.Now().Add(time.Hour * 24)

	// Begin a transaction to ensure both operations complete or fail together
	tx, err := db.Begin()
	if err != nil {
		return "", err
	}

	// Delete any existing sessions for this user
	_, err = tx.Exec("DELETE FROM sessions WHERE user_id = ?", userID)
	if err != nil {
		tx.Rollback()
		return "", err
	}

	// Insert new session into database
	_, err = tx.Exec(
		"INSERT INTO sessions (id, user_id, expires_at) VALUES (?, ?, ?)",
		sessionID, userID, expiresAt,
	)
	if err != nil {
		tx.Rollback()
		return "", err
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return "", err
	}

	return sessionID, nil
}

// ValidateSession checks if a session is valid and returns the associated user ID
func ValidateSession(db *sql.DB, sessionID string) (int64, error) {
	var userID int64
	var expiresAt time.Time

	err := db.QueryRow(
		"SELECT user_id, expires_at FROM sessions WHERE id = ?",
		sessionID,
	).Scan(&userID, &expiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, errors.New("session not found")
		}
		return 0, err
	}

	// Check if session has expired
	if time.Now().After(expiresAt) {
		// Delete expired session
		_, err := db.Exec("DELETE FROM sessions WHERE id = ?", sessionID)
		if err != nil {
			return 0, err
		}
		return 0, errors.New("session expired")
	}

	return userID, nil
}

// DeleteSession removes a session from the database
func DeleteSession(db *sql.DB, sessionID string) error {
	_, err := db.Exec("DELETE FROM sessions WHERE id = ?", sessionID)
	return err
}

// CleanExpiredSessions removes all expired sessions from the database
func CleanExpiredSessions(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM sessions WHERE expires_at < ?", time.Now())
	return err
}
