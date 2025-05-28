package utils

import (
	"database/sql"
	"os"
	"testing"
	"time"

	"forum/database"
)

// setupSessionTestDB creates a temporary database for testing sessions
func setupSessionTestDB(t *testing.T) (*sql.DB, func(), int64) {
	// Create a temporary database file for testing
	tempDBPath := "./test_session.db"

	// Initialize the database
	db, err := database.InitDB(tempDBPath)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Run migrations to create tables
	err = database.RunMigrations(db)
	if err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Create a test user
	result, err := db.Exec(
		"INSERT INTO users (username, email, password) VALUES (?, ?, ?)",
		"testuser", "test@example.com", "hashedpassword",
	)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	userID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("Failed to get user ID: %v", err)
	}

	// Return the database, cleanup function, and user ID
	return db, func() {
		db.Close()
		os.Remove(tempDBPath)
	}, userID
}

func TestCreateSession(t *testing.T) {
	db, cleanup, userID := setupSessionTestDB(t)
	defer cleanup()

	// Test creating a session
	sessionID, err := CreateSession(db, userID)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	if sessionID == "" {
		t.Fatal("Session ID is empty")
	}

	// Verify the session was created in the database
	var dbUserID int64
	var expiresAt time.Time
	err = db.QueryRow(
		"SELECT user_id, expires_at FROM sessions WHERE id = ?",
		sessionID,
	).Scan(&dbUserID, &expiresAt)
	if err != nil {
		t.Fatalf("Failed to query session: %v", err)
	}
	if dbUserID != userID {
		t.Errorf("Expected user ID %d, got %d", userID, dbUserID)
	}
	if expiresAt.Before(time.Now()) {
		t.Error("Session expiration time is in the past")
	}

	// Test creating another session for the same user (should invalidate the first one)
	sessionID2, err := CreateSession(db, userID)
	if err != nil {
		t.Fatalf("Failed to create second session: %v", err)
	}
	if sessionID2 == "" {
		t.Fatal("Second session ID is empty")
	}
	if sessionID2 == sessionID {
		t.Fatal("Second session ID is the same as the first one")
	}

	// Verify the first session was invalidated
	var count int
	err = db.QueryRow(
		"SELECT COUNT(*) FROM sessions WHERE id = ?",
		sessionID,
	).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query first session: %v", err)
	}
	if count != 0 {
		t.Error("First session was not invalidated")
	}

	// Verify the second session exists
	err = db.QueryRow(
		"SELECT COUNT(*) FROM sessions WHERE id = ?",
		sessionID2,
	).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query second session: %v", err)
	}
	if count != 1 {
		t.Error("Second session was not created")
	}
}

func TestValidateSession(t *testing.T) {
	db, cleanup, userID := setupSessionTestDB(t)
	defer cleanup()

	// Create a session
	sessionID, err := CreateSession(db, userID)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Test validating a valid session
	validatedUserID, err := ValidateSession(db, sessionID)
	if err != nil {
		t.Fatalf("Failed to validate session: %v", err)
	}
	if validatedUserID != userID {
		t.Errorf("Expected user ID %d, got %d", userID, validatedUserID)
	}

	// Test validating a non-existent session
	_, err = ValidateSession(db, "non-existent-session-id")
	if err == nil {
		t.Fatal("Expected error for non-existent session, got nil")
	}

	// Test validating an expired session
	// First, create a session with an expiration time in the past
	expiredSessionID := "expired-session-id"
	expiresAt := time.Now().Add(-1 * time.Hour) // 1 hour in the past
	_, err = db.Exec(
		"INSERT INTO sessions (id, user_id, expires_at) VALUES (?, ?, ?)",
		expiredSessionID, userID, expiresAt,
	)
	if err != nil {
		t.Fatalf("Failed to create expired session: %v", err)
	}

	// Now validate the expired session
	_, err = ValidateSession(db, expiredSessionID)
	if err == nil {
		t.Fatal("Expected error for expired session, got nil")
	}

	// Verify the expired session was deleted
	var count int
	err = db.QueryRow(
		"SELECT COUNT(*) FROM sessions WHERE id = ?",
		expiredSessionID,
	).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query expired session: %v", err)
	}
	if count != 0 {
		t.Error("Expired session was not deleted")
	}
}

func TestDeleteSession(t *testing.T) {
	db, cleanup, userID := setupSessionTestDB(t)
	defer cleanup()

	// Create a session
	sessionID, err := CreateSession(db, userID)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Test deleting the session
	err = DeleteSession(db, sessionID)
	if err != nil {
		t.Fatalf("Failed to delete session: %v", err)
	}

	// Verify the session was deleted
	var count int
	err = db.QueryRow(
		"SELECT COUNT(*) FROM sessions WHERE id = ?",
		sessionID,
	).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query deleted session: %v", err)
	}
	if count != 0 {
		t.Error("Session was not deleted")
	}

	// Test deleting a non-existent session (should not error)
	err = DeleteSession(db, "non-existent-session-id")
	if err != nil {
		t.Fatalf("Failed to delete non-existent session: %v", err)
	}
}

func TestCleanExpiredSessions(t *testing.T) {
	db, cleanup, userID := setupSessionTestDB(t)
	defer cleanup()

	// Create an expired session
	expiredSessionID := "expired-session-id"
	expiresAt := time.Now().Add(-1 * time.Hour) // 1 hour in the past
	_, err := db.Exec(
		"INSERT INTO sessions (id, user_id, expires_at) VALUES (?, ?, ?)",
		expiredSessionID, userID, expiresAt,
	)
	if err != nil {
		t.Fatalf("Failed to create expired session: %v", err)
	}

	// Create a valid session
	validSessionID := "valid-session-id"
	validExpiresAt := time.Now().Add(1 * time.Hour) // 1 hour in the future
	_, err = db.Exec(
		"INSERT INTO sessions (id, user_id, expires_at) VALUES (?, ?, ?)",
		validSessionID, userID, validExpiresAt,
	)
	if err != nil {
		t.Fatalf("Failed to create valid session: %v", err)
	}

	// Test cleaning expired sessions
	err = CleanExpiredSessions(db)
	if err != nil {
		t.Fatalf("Failed to clean expired sessions: %v", err)
	}

	// Verify the expired session was deleted
	var count int
	err = db.QueryRow(
		"SELECT COUNT(*) FROM sessions WHERE id = ?",
		expiredSessionID,
	).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query expired session: %v", err)
	}
	if count != 0 {
		t.Error("Expired session was not deleted")
	}

	// Verify the valid session was not deleted
	err = db.QueryRow(
		"SELECT COUNT(*) FROM sessions WHERE id = ?",
		validSessionID,
	).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query valid session: %v", err)
	}
	if count != 1 {
		t.Error("Valid session was deleted")
	}
}
