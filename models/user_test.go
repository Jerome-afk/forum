package models

import (
	"database/sql"
	"os"
	"testing"

	"forum/database"
)

// setupTestDB creates a temporary database for testing
func setupTestDB(t *testing.T) (*sql.DB, func()) {
	// Create a temporary database file for testing
	tempDBPath := "./test_user.db"

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

	// Return the database and a cleanup function
	return db, func() {
		db.Close()
		os.Remove(tempDBPath)
	}
}

func TestCreateUser(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Test creating a user
	userID, err := CreateUser(db, "testuser", "test@example.com", "password123")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	if userID <= 0 {
		t.Fatalf("Expected user ID > 0, got %d", userID)
	}

	// Test duplicate username
	_, err = CreateUser(db, "testuser", "another@example.com", "password123")
	if err == nil {
		t.Fatal("Expected error for duplicate username, got nil")
	}

	// Test duplicate email
	_, err = CreateUser(db, "anotheruser", "test@example.com", "password123")
	if err == nil {
		t.Fatal("Expected error for duplicate email, got nil")
	}
}

func TestGetUserByID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create a test user
	userID, err := CreateUser(db, "testuser", "test@example.com", "password123")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Test getting the user by ID
	user, err := GetUserByID(db, userID)
	if err != nil {
		t.Fatalf("Failed to get user by ID: %v", err)
	}
	if user.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", user.Username)
	}
	if user.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", user.Email)
	}

	// Test getting a non-existent user
	_, err = GetUserByID(db, 9999)
	if err == nil {
		t.Fatal("Expected error for non-existent user, got nil")
	}
}

func TestAuthenticateUser(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create a test user
	_, err := CreateUser(db, "testuser", "test@example.com", "password123")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Test successful authentication
	user, err := AuthenticateUser(db, "test@example.com", "password123")
	if err != nil {
		t.Fatalf("Failed to authenticate user: %v", err)
	}
	if user.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", user.Username)
	}

	// Test authentication with wrong password
	_, err = AuthenticateUser(db, "test@example.com", "wrongpassword")
	if err == nil {
		t.Fatal("Expected error for wrong password, got nil")
	}

	// Test authentication with non-existent email
	_, err = AuthenticateUser(db, "nonexistent@example.com", "password123")
	if err == nil {
		t.Fatal("Expected error for non-existent email, got nil")
	}
}
