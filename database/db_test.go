package database

import (
	"os"
	"testing"
)

func TestInitDB(t *testing.T) {
	// Create a temporary database file for testing
	tempDBPath := "./test_forum.db"

	// Clean up after the test
	defer os.Remove(tempDBPath)

	// Test database initialization
	db, err := InitDB(tempDBPath)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Verify that the database connection is valid
	err = db.Ping()
	if err != nil {
		t.Fatalf("Database connection is not valid: %v", err)
	}

	// Check that the database file was created
	_, err = os.Stat(tempDBPath)
	if os.IsNotExist(err) {
		t.Fatalf("Database file was not created")
	}
}
