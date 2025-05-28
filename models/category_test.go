package models

import (
	"database/sql"
	"os"
	"testing"

	"forum/database"
)

// setupCategoryTestDB creates a temporary database for testing categories
func setupCategoryTestDB(t *testing.T) (*sql.DB, func()) {
	// Create a temporary database file for testing
	tempDBPath := "./test_category.db"

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

func TestCreateCategory(t *testing.T) {
	db, cleanup := setupCategoryTestDB(t)
	defer cleanup()

	// Test creating a category
	categoryID, err := CreateCategory(db, "Test Category")
	if err != nil {
		t.Fatalf("Failed to create category: %v", err)
	}
	if categoryID <= 0 {
		t.Fatalf("Expected category ID > 0, got %d", categoryID)
	}

	// Test creating a duplicate category
	_, err = CreateCategory(db, "Test Category")
	if err == nil {
		t.Fatal("Expected error for duplicate category, got nil")
	}
}

func TestGetAllCategories(t *testing.T) {
	db, cleanup := setupCategoryTestDB(t)
	defer cleanup()

	// Create multiple test categories
	category1ID, err := CreateCategory(db, "Category 1")
	if err != nil {
		t.Fatalf("Failed to create category 1: %v", err)
	}

	category2ID, err := CreateCategory(db, "Category 2")
	if err != nil {
		t.Fatalf("Failed to create category 2: %v", err)
	}

	category3ID, err := CreateCategory(db, "Category 3")
	if err != nil {
		t.Fatalf("Failed to create category 3: %v", err)
	}

	// Test getting all categories
	categories, err := GetAllCategories(db)
	if err != nil {
		t.Fatalf("Failed to get categories: %v", err)
	}
	if len(categories) != 3 {
		t.Errorf("Expected 3 categories, got %d", len(categories))
	}

	// Verify category names and IDs
	categoryMap := make(map[string]int64)
	for _, category := range categories {
		categoryMap[category.Name] = category.ID
	}

	if id, ok := categoryMap["Category 1"]; !ok || id != category1ID {
		t.Errorf("Category 1 not found or has wrong ID")
	}
	if id, ok := categoryMap["Category 2"]; !ok || id != category2ID {
		t.Errorf("Category 2 not found or has wrong ID")
	}
	if id, ok := categoryMap["Category 3"]; !ok || id != category3ID {
		t.Errorf("Category 3 not found or has wrong ID")
	}
}

func TestGetCategoryByID(t *testing.T) {
	db, cleanup := setupCategoryTestDB(t)
	defer cleanup()

	// Create a test category
	categoryID, err := CreateCategory(db, "Test Category")
	if err != nil {
		t.Fatalf("Failed to create category: %v", err)
	}

	// Test getting the category by ID
	category, err := GetCategoryByID(db, categoryID)
	if err != nil {
		t.Fatalf("Failed to get category by ID: %v", err)
	}
	if category.Name != "Test Category" {
		t.Errorf("Expected name 'Test Category', got '%s'", category.Name)
	}
	if category.ID != categoryID {
		t.Errorf("Expected ID %d, got %d", categoryID, category.ID)
	}

	// Test getting a non-existent category
	_, err = GetCategoryByID(db, 9999)
	if err == nil {
		t.Fatal("Expected error for non-existent category, got nil")
	}
}

func TestCategoryExists(t *testing.T) {
	db, cleanup := setupCategoryTestDB(t)
	defer cleanup()

	// Create a test category
	_, err := CreateCategory(db, "Test Category")
	if err != nil {
		t.Fatalf("Failed to create category: %v", err)
	}

	// Test checking if a category exists
	exists, err := CategoryExists(db, "Test Category")
	if err != nil {
		t.Fatalf("Failed to check if category exists: %v", err)
	}
	if !exists {
		t.Error("Expected category to exist, but it doesn't")
	}

	// Test checking if a non-existent category exists
	exists, err = CategoryExists(db, "Non-existent Category")
	if err != nil {
		t.Fatalf("Failed to check if category exists: %v", err)
	}
	if exists {
		t.Error("Expected category to not exist, but it does")
	}
}
