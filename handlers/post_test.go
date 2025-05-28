package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"

	"forum/database"
	"forum/models"
)

// setupPostTestDB creates a temporary database for testing post handlers
func setupPostTestDB(t *testing.T) (*sql.DB, func()) {
	// Create a temporary database file for testing
	tempDBPath := "./test_post.db"

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

// setupTestUser creates a test user and returns the user ID
func setupTestUser(t *testing.T, db *sql.DB) int64 {
	userID, err := models.CreateUser(db, "testuser", "test@example.com", "password123")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	return userID
}

// setupTestCategories creates test categories and returns their IDs
func setupTestCategories(t *testing.T, db *sql.DB) []int64 {
	categoryIDs := make([]int64, 0, 2)

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

	return categoryIDs
}

// setupTestPost creates a test post and returns the post ID
func setupTestPost(t *testing.T, db *sql.DB, userID int64, categoryIDs []int64) int64 {
	postID, err := models.CreatePost(db, "Test Post Title", "This is a test post content.", userID, categoryIDs)
	if err != nil {
		t.Fatalf("Failed to create test post: %v", err)
	}
	return postID
}

// createAuthenticatedRequest creates an HTTP request with the database in context and a user session
func createAuthenticatedRequest(method, url string, body *bytes.Buffer, db *sql.DB, userID int64) *http.Request {
	var req *http.Request
	if body != nil {
		req, _ = http.NewRequest(method, url, body)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else {
		req, _ = http.NewRequest(method, url, nil)
	}

	// Add database to context
	ctx := context.WithValue(req.Context(), dbContextKey, db)

	// Add user to context
	user, err := models.GetUserByID(db, userID)
	if err == nil {
		ctx = context.WithValue(ctx, userContextKey, user)
	}

	return req.WithContext(ctx)
}

func TestGetPostHandler(t *testing.T) {
	db, cleanup := setupPostTestDB(t)
	defer cleanup()

	// Setup test data
	userID := setupTestUser(t, db)
	categoryIDs := setupTestCategories(t, db)
	postID := setupTestPost(t, db, userID, categoryIDs)

	// Create a comment on the post
	_, err := models.CreateComment(db, "This is a test comment.", userID, postID)
	if err != nil {
		t.Fatalf("Failed to create test comment: %v", err)
	}

	// Test GET request to view a post
	req := createRequestWithDB("GET", "/post?id="+strconv.FormatInt(postID, 10), nil, db)
	rr := httptest.NewRecorder()

	// Add post ID to request context
	ctx := context.WithValue(req.Context(), "postID", postID)
	req = req.WithContext(ctx)

	GetPostHandler(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check that the response contains the post title and content
	response := rr.Body.String()
	if !strings.Contains(response, "Test Post Title") {
		t.Errorf("Expected response to contain post title, got: %s", response)
	}
	if !strings.Contains(response, "This is a test post content.") {
		t.Errorf("Expected response to contain post content, got: %s", response)
	}
	if !strings.Contains(response, "This is a test comment.") {
		t.Errorf("Expected response to contain comment, got: %s", response)
	}

	// Test with non-existent post ID
	req = createRequestWithDB("GET", "/post?id=999", nil, db)
	rr = httptest.NewRecorder()

	// Add invalid post ID to request context
	ctx = context.WithValue(req.Context(), "postID", int64(999))
	req = req.WithContext(ctx)

	GetPostHandler(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}
}

func TestCreatePostHandler(t *testing.T) {
	db, cleanup := setupPostTestDB(t)
	defer cleanup()

	// Setup test data
	userID := setupTestUser(t, db)
	_ = setupTestCategories(t, db)

	// Test GET request to create post page
	req := createAuthenticatedRequest("GET", "/create-post", nil, db, userID)
	rr := httptest.NewRecorder()

	CreatePostHandler(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Test successful post creation
	formData := url.Values{}
	formData.Set("title", "New Test Post")
	formData.Set("content", "This is a new test post content.")
	formData.Set("category", "1")
	formData.Add("category", "2")

	req = createAuthenticatedRequest("POST", "/create-post", bytes.NewBufferString(formData.Encode()), db, userID)
	rr = httptest.NewRecorder()

	CreatePostHandler(rr, req)

	if status := rr.Code; status != http.StatusSeeOther {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusSeeOther)
	}

	// Verify the post was created in the database
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM posts WHERE title = ? AND content = ?", "New Test Post", "This is a new test post content.").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query post: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 post, got %d", count)
	}

	// Test post creation with missing title
	formData = url.Values{}
	formData.Set("title", "")
	formData.Set("content", "This is a test post content.")
	formData.Set("category", "1")

	req = createAuthenticatedRequest("POST", "/create-post", bytes.NewBufferString(formData.Encode()), db, userID)
	rr = httptest.NewRecorder()

	CreatePostHandler(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	if !strings.Contains(rr.Body.String(), "Title is required") {
		t.Errorf("Expected error message about missing title, got: %s", rr.Body.String())
	}

	// Test post creation with missing content
	formData = url.Values{}
	formData.Set("title", "Test Post Title")
	formData.Set("content", "")
	formData.Set("category", "1")

	req = createAuthenticatedRequest("POST", "/create-post", bytes.NewBufferString(formData.Encode()), db, userID)
	rr = httptest.NewRecorder()

	CreatePostHandler(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	if !strings.Contains(rr.Body.String(), "Content is required") {
		t.Errorf("Expected error message about missing content, got: %s", rr.Body.String())
	}

	// Test post creation with missing category
	formData = url.Values{}
	formData.Set("title", "Test Post Title")
	formData.Set("content", "This is a test post content.")

	req = createAuthenticatedRequest("POST", "/create-post", bytes.NewBufferString(formData.Encode()), db, userID)
	rr = httptest.NewRecorder()

	CreatePostHandler(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	if !strings.Contains(rr.Body.String(), "At least one category is required") {
		t.Errorf("Expected error message about missing category, got: %s", rr.Body.String())
	}
}

func TestMyPostsHandler(t *testing.T) {
	db, cleanup := setupPostTestDB(t)
	defer cleanup()

	// Setup test data
	userID := setupTestUser(t, db)
	categoryIDs := setupTestCategories(t, db)
	_ = setupTestPost(t, db, userID, categoryIDs)

	// Test GET request to view user's posts
	req := createAuthenticatedRequest("GET", "/my-posts", nil, db, userID)
	rr := httptest.NewRecorder()

	MyPostsHandler(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check that the response contains the post title
	response := rr.Body.String()
	if !strings.Contains(response, "Test Post Title") {
		t.Errorf("Expected response to contain post title, got: %s", response)
	}

	// Test without authentication
	req = createRequestWithDB("GET", "/my-posts", nil, db)
	rr = httptest.NewRecorder()

	MyPostsHandler(rr, req)

	if status := rr.Code; status != http.StatusSeeOther {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusSeeOther)
	}
}

func TestLikedPostsHandler(t *testing.T) {
	db, cleanup := setupPostTestDB(t)
	defer cleanup()

	// Setup test data
	userID := setupTestUser(t, db)
	categoryIDs := setupTestCategories(t, db)
	postID := setupTestPost(t, db, userID, categoryIDs)

	// Like the post
	err := models.ReactToPost(db, postID, userID, 1) // 1 for like
	if err != nil {
		t.Fatalf("Failed to react to post: %v", err)
	}

	// Test GET request to view liked posts
	req := createAuthenticatedRequest("GET", "/liked-posts", nil, db, userID)
	rr := httptest.NewRecorder()

	LikedPostsHandler(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check that the response contains the post title
	response := rr.Body.String()
	if !strings.Contains(response, "Test Post Title") {
		t.Errorf("Expected response to contain post title, got: %s", response)
	}

	// Test without authentication
	req = createRequestWithDB("GET", "/liked-posts", nil, db)
	rr = httptest.NewRecorder()

	LikedPostsHandler(rr, req)

	if status := rr.Code; status != http.StatusSeeOther {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusSeeOther)
	}
}

func TestReactPostHandler(t *testing.T) {
	db, cleanup := setupPostTestDB(t)
	defer cleanup()

	// Setup test data
	userID := setupTestUser(t, db)
	categoryIDs := setupTestCategories(t, db)
	postID := setupTestPost(t, db, userID, categoryIDs)

	// Test liking a post
	formData := url.Values{}
	formData.Set("post_id", strconv.FormatInt(postID, 10))
	formData.Set("reaction", "1") // 1 for like

	req := createAuthenticatedRequest("POST", "/react-post", bytes.NewBufferString(formData.Encode()), db, userID)
	rr := httptest.NewRecorder()

	ReactPostHandler(rr, req)

	if status := rr.Code; status != http.StatusSeeOther {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusSeeOther)
	}

	// Verify the reaction was recorded in the database
	var reaction int
	err := db.QueryRow("SELECT reaction FROM post_reactions WHERE post_id = ? AND user_id = ?", postID, userID).Scan(&reaction)
	if err != nil {
		t.Fatalf("Failed to query reaction: %v", err)
	}
	if reaction != 1 {
		t.Errorf("Expected reaction 1 (like), got %d", reaction)
	}

	// Test disliking the same post (changing reaction)
	formData = url.Values{}
	formData.Set("post_id", strconv.FormatInt(postID, 10))
	formData.Set("reaction", "-1") // -1 for dislike

	req = createAuthenticatedRequest("POST", "/react-post", bytes.NewBufferString(formData.Encode()), db, userID)
	rr = httptest.NewRecorder()

	ReactPostHandler(rr, req)

	if status := rr.Code; status != http.StatusSeeOther {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusSeeOther)
	}

	// Verify the reaction was updated in the database
	err = db.QueryRow("SELECT reaction FROM post_reactions WHERE post_id = ? AND user_id = ?", postID, userID).Scan(&reaction)
	if err != nil {
		t.Fatalf("Failed to query reaction: %v", err)
	}
	if reaction != -1 {
		t.Errorf("Expected reaction -1 (dislike), got %d", reaction)
	}

	// Test removing reaction by submitting the same reaction again
	formData = url.Values{}
	formData.Set("post_id", strconv.FormatInt(postID, 10))
	formData.Set("reaction", "-1") // -1 for dislike (same as current)

	req = createAuthenticatedRequest("POST", "/react-post", bytes.NewBufferString(formData.Encode()), db, userID)
	rr = httptest.NewRecorder()

	ReactPostHandler(rr, req)

	if status := rr.Code; status != http.StatusSeeOther {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusSeeOther)
	}

	// Verify the reaction was removed from the database
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM post_reactions WHERE post_id = ? AND user_id = ?", postID, userID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query reaction count: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected reaction to be removed, but it still exists")
	}

	// Test without authentication
	formData = url.Values{}
	formData.Set("post_id", strconv.FormatInt(postID, 10))
	formData.Set("reaction", "1")

	req = createRequestWithDB("POST", "/react-post", bytes.NewBufferString(formData.Encode()), db)
	rr = httptest.NewRecorder()

	ReactPostHandler(rr, req)

	if status := rr.Code; status != http.StatusSeeOther {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusSeeOther)
	}
	// Should redirect to login page
	if location := rr.Header().Get("Location"); location != "/login" {
		t.Errorf("Expected redirect to /login, got: %s", location)
	}
}
