package models

import (
	"database/sql"
	"os"
	"testing"

	"forum/database"
)

// setupPostTestDB creates a temporary database for testing posts
func setupPostTestDB(t *testing.T) (*sql.DB, func(), int64) {
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

	// Create a test user
	userID, err := CreateUser(db, "testuser", "test@example.com", "password123")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Return the database, cleanup function, and user ID
	return db, func() {
		db.Close()
		os.Remove(tempDBPath)
	}, userID
}

func TestCreatePost(t *testing.T) {
	db, cleanup, userID := setupPostTestDB(t)
	defer cleanup()

	// Create a test category
	categoryID, err := CreateCategory(db, "Test Category")
	if err != nil {
		t.Fatalf("Failed to create test category: %v", err)
	}

	// Test creating a post with a category
	postID, err := CreatePost(db, "Test Post", "This is a test post content.", userID, []int64{categoryID})
	if err != nil {
		t.Fatalf("Failed to create post: %v", err)
	}
	if postID <= 0 {
		t.Fatalf("Expected post ID > 0, got %d", postID)
	}

	// Test creating a post without categories
	postID2, err := CreatePost(db, "Another Test Post", "This is another test post content.", userID, []int64{})
	if err != nil {
		t.Fatalf("Failed to create post without categories: %v", err)
	}
	if postID2 <= 0 {
		t.Fatalf("Expected post ID > 0, got %d", postID2)
	}
}

func TestGetPostByID(t *testing.T) {
	db, cleanup, userID := setupPostTestDB(t)
	defer cleanup()

	// Create a test category
	categoryID, err := CreateCategory(db, "Test Category")
	if err != nil {
		t.Fatalf("Failed to create test category: %v", err)
	}

	// Create a test post
	postID, err := CreatePost(db, "Test Post", "This is a test post content.", userID, []int64{categoryID})
	if err != nil {
		t.Fatalf("Failed to create test post: %v", err)
	}

	// Test getting the post by ID
	post, err := GetPostByID(db, postID, userID)
	if err != nil {
		t.Fatalf("Failed to get post by ID: %v", err)
	}
	if post.Title != "Test Post" {
		t.Errorf("Expected title 'Test Post', got '%s'", post.Title)
	}
	if post.Content != "This is a test post content." {
		t.Errorf("Expected content 'This is a test post content.', got '%s'", post.Content)
	}
	if post.UserID != userID {
		t.Errorf("Expected user ID %d, got %d", userID, post.UserID)
	}
	if len(post.Categories) != 1 {
		t.Errorf("Expected 1 category, got %d", len(post.Categories))
	} else if post.Categories[0].ID != categoryID {
		t.Errorf("Expected category ID %d, got %d", categoryID, post.Categories[0].ID)
	}

	// Test getting a non-existent post
	_, err = GetPostByID(db, 9999, userID)
	if err == nil {
		t.Fatal("Expected error for non-existent post, got nil")
	}
}

func TestGetPosts(t *testing.T) {
	db, cleanup, userID := setupPostTestDB(t)
	defer cleanup()

	// Create test categories
	category1ID, err := CreateCategory(db, "Category 1")
	if err != nil {
		t.Fatalf("Failed to create test category 1: %v", err)
	}
	category2ID, err := CreateCategory(db, "Category 2")
	if err != nil {
		t.Fatalf("Failed to create test category 2: %v", err)
	}

	// Create test posts
	post1ID, err := CreatePost(db, "Post 1", "Content 1", userID, []int64{category1ID})
	if err != nil {
		t.Fatalf("Failed to create test post 1: %v", err)
	}
	_, err = CreatePost(db, "Post 2", "Content 2", userID, []int64{category2ID})
	if err != nil {
		t.Fatalf("Failed to create test post 2: %v", err)
	}
	_, err = CreatePost(db, "Post 3", "Content 3", userID, []int64{category1ID, category2ID})
	if err != nil {
		t.Fatalf("Failed to create test post 3: %v", err)
	}

	// Test getting all posts
	posts, err := GetPosts(db, userID, 0, 0, false)
	if err != nil {
		t.Fatalf("Failed to get posts: %v", err)
	}
	if len(posts) != 3 {
		t.Errorf("Expected 3 posts, got %d", len(posts))
	}

	// Test filtering by category
	posts, err = GetPosts(db, userID, category1ID, 0, false)
	if err != nil {
		t.Fatalf("Failed to get posts filtered by category: %v", err)
	}
	if len(posts) != 2 {
		t.Errorf("Expected 2 posts for category 1, got %d", len(posts))
	}

	// Test filtering by user
	posts, err = GetPosts(db, userID, 0, userID, false)
	if err != nil {
		t.Fatalf("Failed to get posts filtered by user: %v", err)
	}
	if len(posts) != 3 {
		t.Errorf("Expected 3 posts for user, got %d", len(posts))
	}

	// Test reaction functionality
	err = ReactToPost(db, post1ID, userID, 1) // Like post 1
	if err != nil {
		t.Fatalf("Failed to react to post: %v", err)
	}

	// Test getting liked posts
	posts, err = GetPosts(db, userID, 0, 0, true)
	if err != nil {
		t.Fatalf("Failed to get liked posts: %v", err)
	}
	if len(posts) != 1 {
		t.Errorf("Expected 1 liked post, got %d", len(posts))
	}
}

func TestReactToPost(t *testing.T) {
	db, cleanup, userID := setupPostTestDB(t)
	defer cleanup()

	// Create a test post
	postID, err := CreatePost(db, "Test Post", "This is a test post content.", userID, []int64{})
	if err != nil {
		t.Fatalf("Failed to create test post: %v", err)
	}

	// Test liking a post
	err = ReactToPost(db, postID, userID, 1)
	if err != nil {
		t.Fatalf("Failed to like post: %v", err)
	}

	// Verify the reaction
	post, err := GetPostByID(db, postID, userID)
	if err != nil {
		t.Fatalf("Failed to get post: %v", err)
	}
	if post.Likes != 1 {
		t.Errorf("Expected 1 like, got %d", post.Likes)
	}
	if post.UserReaction != 1 {
		t.Errorf("Expected user reaction 1, got %d", post.UserReaction)
	}

	// Test changing reaction to dislike
	err = ReactToPost(db, postID, userID, -1)
	if err != nil {
		t.Fatalf("Failed to change reaction: %v", err)
	}

	// Verify the changed reaction
	post, err = GetPostByID(db, postID, userID)
	if err != nil {
		t.Fatalf("Failed to get post: %v", err)
	}
	if post.Likes != 0 {
		t.Errorf("Expected 0 likes, got %d", post.Likes)
	}
	if post.Dislikes != 1 {
		t.Errorf("Expected 1 dislike, got %d", post.Dislikes)
	}
	if post.UserReaction != -1 {
		t.Errorf("Expected user reaction -1, got %d", post.UserReaction)
	}

	// Test removing reaction by clicking the same button
	err = ReactToPost(db, postID, userID, -1)
	if err != nil {
		t.Fatalf("Failed to remove reaction: %v", err)
	}

	// Verify the reaction was removed
	post, err = GetPostByID(db, postID, userID)
	if err != nil {
		t.Fatalf("Failed to get post: %v", err)
	}
	if post.Dislikes != 0 {
		t.Errorf("Expected 0 dislikes, got %d", post.Dislikes)
	}
	if post.UserReaction != 0 {
		t.Errorf("Expected user reaction 0, got %d", post.UserReaction)
	}
}
