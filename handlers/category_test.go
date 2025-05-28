package handlers

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"forum/database"
	"forum/models"
)

// setupCategoryTestDB creates a temporary database for testing category handlers
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

// setupCategoryTestData creates test categories
func setupCategoryTestData(t *testing.T, db *sql.DB) []int64 {
	categoryIDs := make([]int64, 0, 3)

	// Create test categories
	catID1, err := models.CreateCategory(db, "Technology")
	if err != nil {
		t.Fatalf("Failed to create test category: %v", err)
	}
	categoryIDs = append(categoryIDs, catID1)

	catID2, err := models.CreateCategory(db, "Science")
	if err != nil {
		t.Fatalf("Failed to create test category: %v", err)
	}
	categoryIDs = append(categoryIDs, catID2)

	catID3, err := models.CreateCategory(db, "Art")
	if err != nil {
		t.Fatalf("Failed to create test category: %v", err)
	}
	categoryIDs = append(categoryIDs, catID3)

	return categoryIDs
}

func TestHomeHandler(t *testing.T) {
	db, cleanup := setupCategoryTestDB(t)
	defer cleanup()

	// Setup test data
	categoryIDs := setupCategoryTestData(t, db)

	// Create a test user
	userID, err := models.CreateUser(db, "testuser", "test@example.com", "password123")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create some test posts
	_, err = models.CreatePost(db, "First Test Post", "This is the first test post content.", userID, categoryIDs[:1])
	if err != nil {
		t.Fatalf("Failed to create test post: %v", err)
	}

	_, err = models.CreatePost(db, "Second Test Post", "This is the second test post content.", userID, categoryIDs[1:2])
	if err != nil {
		t.Fatalf("Failed to create test post: %v", err)
	}

	// Test GET request to home page
	req := createRequestWithDB("GET", "/", nil, db)
	rr := httptest.NewRecorder()

	HomeHandler(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check that the response contains the post titles and categories
	response := rr.Body.String()
	if !strings.Contains(response, "First Test Post") {
		t.Errorf("Expected response to contain first post title, got: %s", response)
	}
	if !strings.Contains(response, "Second Test Post") {
		t.Errorf("Expected response to contain second post title, got: %s", response)
	}
	if !strings.Contains(response, "Technology") {
		t.Errorf("Expected response to contain Technology category, got: %s", response)
	}
	if !strings.Contains(response, "Science") {
		t.Errorf("Expected response to contain Science category, got: %s", response)
	}
	if !strings.Contains(response, "Art") {
		t.Errorf("Expected response to contain Art category, got: %s", response)
	}
}

func TestCategoriesHandler(t *testing.T) {
	db, cleanup := setupCategoryTestDB(t)
	defer cleanup()

	// Setup test data
	setupCategoryTestData(t, db)

	// Test GET request to categories page
	req := createRequestWithDB("GET", "/categories", nil, db)
	rr := httptest.NewRecorder()

	CategoriesHandler(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check that the response contains all categories
	response := rr.Body.String()
	if !strings.Contains(response, "Technology") {
		t.Errorf("Expected response to contain Technology category, got: %s", response)
	}
	if !strings.Contains(response, "Science") {
		t.Errorf("Expected response to contain Science category, got: %s", response)
	}
	if !strings.Contains(response, "Art") {
		t.Errorf("Expected response to contain Art category, got: %s", response)
	}
}

func TestCategoryPostsHandler(t *testing.T) {
	db, cleanup := setupCategoryTestDB(t)
	defer cleanup()

	// Setup test data
	categoryIDs := setupCategoryTestData(t, db)

	// Create a test user
	userID, err := models.CreateUser(db, "testuser", "test@example.com", "password123")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create test posts in different categories
	_, err = models.CreatePost(db, "Technology Post", "This is a technology post.", userID, categoryIDs[:1])
	if err != nil {
		t.Fatalf("Failed to create test post: %v", err)
	}

	_, err = models.CreatePost(db, "Science Post", "This is a science post.", userID, categoryIDs[1:2])
	if err != nil {
		t.Fatalf("Failed to create test post: %v", err)
	}

	// Test GET request to view posts in the Technology category
	req := createRequestWithDB("GET", "/category?id=1", nil, db)
	rr := httptest.NewRecorder()

	// Add category ID to request context
	ctx := context.WithValue(req.Context(), "categoryID", categoryIDs[0])
	req = req.WithContext(ctx)

	CategoryPostsHandler(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check that the response contains the Technology post but not the Science post
	response := rr.Body.String()
	if !strings.Contains(response, "Technology Post") {
		t.Errorf("Expected response to contain Technology post, got: %s", response)
	}
	if strings.Contains(response, "Science Post") {
		t.Errorf("Expected response to NOT contain Science post, got: %s", response)
	}

	// Test GET request to view posts in the Science category
	req = createRequestWithDB("GET", "/category?id=2", nil, db)
	rr = httptest.NewRecorder()

	// Add category ID to request context
	ctx = context.WithValue(req.Context(), "categoryID", categoryIDs[1])
	req = req.WithContext(ctx)

	CategoryPostsHandler(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check that the response contains the Science post but not the Technology post
	response = rr.Body.String()
	if !strings.Contains(response, "Science Post") {
		t.Errorf("Expected response to contain Science post, got: %s", response)
	}
	if strings.Contains(response, "Technology Post") {
		t.Errorf("Expected response to NOT contain Technology post, got: %s", response)
	}

	// Test with non-existent category ID
	req = createRequestWithDB("GET", "/category?id=999", nil, db)
	rr = httptest.NewRecorder()

	// Add invalid category ID to request context
	ctx = context.WithValue(req.Context(), "categoryID", int64(999))
	req = req.WithContext(ctx)

	CategoryPostsHandler(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}
}
