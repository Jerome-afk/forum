package handlers

import (
	"bytes"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"forum/database"
	"forum/models"
)

// setupCommentTestDB creates a temporary database for testing comment handlers
func setupCommentTestDB(t *testing.T) (*sql.DB, func()) {
	// Create a temporary database file for testing
	tempDBPath := "./test_comment.db"

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

// setupCommentTestData creates test users, categories, posts, and comments
func setupCommentTestData(t *testing.T, db *sql.DB) (int64, int64, int64) {
	// Create a test user
	userID, err := models.CreateUser(db, "testuser", "test@example.com", "password123")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create a test category
	categoryID, err := models.CreateCategory(db, "Test Category")
	if err != nil {
		t.Fatalf("Failed to create test category: %v", err)
	}

	// Create a test post
	postID, err := models.CreatePost(db, "Test Post", "This is a test post content.", userID, []int64{categoryID})
	if err != nil {
		t.Fatalf("Failed to create test post: %v", err)
	}

	return userID, categoryID, postID
}

func TestCreateCommentHandler(t *testing.T) {
	db, cleanup := setupCommentTestDB(t)
	defer cleanup()

	// Setup test data
	userID, _, postID := setupCommentTestData(t, db)

	// Test successful comment creation
	formData := url.Values{}
	formData.Set("content", "This is a test comment.")
	formData.Set("post_id", string(postID))

	req := createAuthenticatedRequest("POST", "/comment", bytes.NewBufferString(formData.Encode()), db, userID)
	rr := httptest.NewRecorder()

	CreateCommentHandler(rr, req)

	if status := rr.Code; status != http.StatusSeeOther {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusSeeOther)
	}

	// Verify the comment was created in the database
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM comments WHERE content = ? AND post_id = ? AND user_id = ?",
		"This is a test comment.", postID, userID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query comment: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 comment, got %d", count)
	}

	// Test comment creation with empty content
	formData = url.Values{}
	formData.Set("content", "")
	formData.Set("post_id", string(postID))

	req = createAuthenticatedRequest("POST", "/comment", bytes.NewBufferString(formData.Encode()), db, userID)
	rr = httptest.NewRecorder()

	CreateCommentHandler(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	if !strings.Contains(rr.Body.String(), "Comment content is required") {
		t.Errorf("Expected error message about empty content, got: %s", rr.Body.String())
	}

	// Test without authentication
	formData = url.Values{}
	formData.Set("content", "This is another test comment.")
	formData.Set("post_id", string(postID))

	req = createRequestWithDB("POST", "/comment", bytes.NewBufferString(formData.Encode()), db)
	rr = httptest.NewRecorder()

	CreateCommentHandler(rr, req)

	if status := rr.Code; status != http.StatusSeeOther {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusSeeOther)
	}
	// Should redirect to login page
	if location := rr.Header().Get("Location"); location != "/login" {
		t.Errorf("Expected redirect to /login, got: %s", location)
	}
}

func TestReactCommentHandler(t *testing.T) {
	db, cleanup := setupCommentTestDB(t)
	defer cleanup()

	// Setup test data
	userID, _, postID := setupCommentTestData(t, db)

	// Create a test comment
	commentID, err := models.CreateComment(db, "This is a test comment.", userID, postID)
	if err != nil {
		t.Fatalf("Failed to create test comment: %v", err)
	}

	// Test liking a comment
	formData := url.Values{}
	formData.Set("comment_id", string(commentID))
	formData.Set("reaction", "1") // 1 for like
	formData.Set("post_id", string(postID))

	req := createAuthenticatedRequest("POST", "/react-comment", bytes.NewBufferString(formData.Encode()), db, userID)
	rr := httptest.NewRecorder()

	ReactCommentHandler(rr, req)

	if status := rr.Code; status != http.StatusSeeOther {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusSeeOther)
	}

	// Verify the reaction was recorded in the database
	var reaction int
	err = db.QueryRow("SELECT reaction FROM comment_reactions WHERE comment_id = ? AND user_id = ?",
		commentID, userID).Scan(&reaction)
	if err != nil {
		t.Fatalf("Failed to query reaction: %v", err)
	}
	if reaction != 1 {
		t.Errorf("Expected reaction 1 (like), got %d", reaction)
	}

	// Test disliking the same comment (changing reaction)
	formData = url.Values{}
	formData.Set("comment_id", string(commentID))
	formData.Set("reaction", "-1") // -1 for dislike
	formData.Set("post_id", string(postID))

	req = createAuthenticatedRequest("POST", "/react-comment", bytes.NewBufferString(formData.Encode()), db, userID)
	rr = httptest.NewRecorder()

	ReactCommentHandler(rr, req)

	if status := rr.Code; status != http.StatusSeeOther {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusSeeOther)
	}

	// Verify the reaction was updated in the database
	err = db.QueryRow("SELECT reaction FROM comment_reactions WHERE comment_id = ? AND user_id = ?",
		commentID, userID).Scan(&reaction)
	if err != nil {
		t.Fatalf("Failed to query reaction: %v", err)
	}
	if reaction != -1 {
		t.Errorf("Expected reaction -1 (dislike), got %d", reaction)
	}

	// Test removing reaction by submitting the same reaction again
	formData = url.Values{}
	formData.Set("comment_id", string(commentID))
	formData.Set("reaction", "-1") // -1 for dislike (same as current)
	formData.Set("post_id", string(postID))

	req = createAuthenticatedRequest("POST", "/react-comment", bytes.NewBufferString(formData.Encode()), db, userID)
	rr = httptest.NewRecorder()

	ReactCommentHandler(rr, req)

	if status := rr.Code; status != http.StatusSeeOther {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusSeeOther)
	}

	// Verify the reaction was removed from the database
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM comment_reactions WHERE comment_id = ? AND user_id = ?",
		commentID, userID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query reaction count: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected reaction to be removed, but it still exists")
	}

	// Test without authentication
	formData = url.Values{}
	formData.Set("comment_id", string(commentID))
	formData.Set("reaction", "1")
	formData.Set("post_id", string(postID))

	req = createRequestWithDB("POST", "/react-comment", bytes.NewBufferString(formData.Encode()), db)
	rr = httptest.NewRecorder()

	ReactCommentHandler(rr, req)

	if status := rr.Code; status != http.StatusSeeOther {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusSeeOther)
	}
	// Should redirect to login page
	if location := rr.Header().Get("Location"); location != "/login" {
		t.Errorf("Expected redirect to /login, got: %s", location)
	}
}
