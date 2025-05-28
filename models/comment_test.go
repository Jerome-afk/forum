package models

import (
	"database/sql"
	"os"
	"testing"

	"forum/database"
)

// setupCommentTestDB creates a temporary database for testing comments
func setupCommentTestDB(t *testing.T) (*sql.DB, func(), int64, int64) {
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

	// Create a test user
	userID, err := CreateUser(db, "testuser", "test@example.com", "password123")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create a test post
	postID, err := CreatePost(db, "Test Post", "This is a test post content.", userID, []int64{})
	if err != nil {
		t.Fatalf("Failed to create test post: %v", err)
	}

	// Return the database, cleanup function, user ID, and post ID
	return db, func() {
		db.Close()
		os.Remove(tempDBPath)
	}, userID, postID
}

func TestCreateComment(t *testing.T) {
	db, cleanup, userID, postID := setupCommentTestDB(t)
	defer cleanup()

	// Test creating a comment
	commentID, err := CreateComment(db, "This is a test comment.", userID, postID)
	if err != nil {
		t.Fatalf("Failed to create comment: %v", err)
	}
	if commentID <= 0 {
		t.Fatalf("Expected comment ID > 0, got %d", commentID)
	}
}

func TestGetCommentsByPostID(t *testing.T) {
	db, cleanup, userID, postID := setupCommentTestDB(t)
	defer cleanup()

	// Create multiple test comments
	_, err := CreateComment(db, "Comment 1", userID, postID)
	if err != nil {
		t.Fatalf("Failed to create comment 1: %v", err)
	}

	_, err = CreateComment(db, "Comment 2", userID, postID)
	if err != nil {
		t.Fatalf("Failed to create comment 2: %v", err)
	}

	// Test getting comments for a post
	comments, err := GetCommentsByPostID(db, postID, userID)
	if err != nil {
		t.Fatalf("Failed to get comments: %v", err)
	}
	if len(comments) != 2 {
		t.Errorf("Expected 2 comments, got %d", len(comments))
	}

	// Verify comment content
	commentContents := map[string]bool{
		"Comment 1": false,
		"Comment 2": false,
	}

	for _, comment := range comments {
		commentContents[comment.Content] = true
		if comment.UserID != userID {
			t.Errorf("Expected user ID %d, got %d", userID, comment.UserID)
		}
		if comment.PostID != postID {
			t.Errorf("Expected post ID %d, got %d", postID, comment.PostID)
		}
	}

	for content, found := range commentContents {
		if !found {
			t.Errorf("Comment with content '%s' not found", content)
		}
	}

	// Test getting comments for a non-existent post
	comments, err = GetCommentsByPostID(db, 9999, userID)
	if err != nil {
		t.Fatalf("Failed to get comments for non-existent post: %v", err)
	}
	if len(comments) != 0 {
		t.Errorf("Expected 0 comments for non-existent post, got %d", len(comments))
	}
}

func TestReactToComment(t *testing.T) {
	db, cleanup, userID, postID := setupCommentTestDB(t)
	defer cleanup()

	// Create a test comment
	commentID, err := CreateComment(db, "This is a test comment.", userID, postID)
	if err != nil {
		t.Fatalf("Failed to create comment: %v", err)
	}

	// Test liking a comment
	err = ReactToComment(db, commentID, userID, 1)
	if err != nil {
		t.Fatalf("Failed to like comment: %v", err)
	}

	// Verify the reaction
	comments, err := GetCommentsByPostID(db, postID, userID)
	if err != nil {
		t.Fatalf("Failed to get comments: %v", err)
	}
	if len(comments) != 1 {
		t.Fatalf("Expected 1 comment, got %d", len(comments))
	}

	comment := comments[0]
	if comment.Likes != 1 {
		t.Errorf("Expected 1 like, got %d", comment.Likes)
	}
	if comment.UserReaction != 1 {
		t.Errorf("Expected user reaction 1, got %d", comment.UserReaction)
	}

	// Test changing reaction to dislike
	err = ReactToComment(db, commentID, userID, -1)
	if err != nil {
		t.Fatalf("Failed to change reaction: %v", err)
	}

	// Verify the changed reaction
	comments, err = GetCommentsByPostID(db, postID, userID)
	if err != nil {
		t.Fatalf("Failed to get comments: %v", err)
	}
	comment = comments[0]
	if comment.Likes != 0 {
		t.Errorf("Expected 0 likes, got %d", comment.Likes)
	}
	if comment.Dislikes != 1 {
		t.Errorf("Expected 1 dislike, got %d", comment.Dislikes)
	}
	if comment.UserReaction != -1 {
		t.Errorf("Expected user reaction -1, got %d", comment.UserReaction)
	}

	// Test removing reaction by clicking the same button
	err = ReactToComment(db, commentID, userID, -1)
	if err != nil {
		t.Fatalf("Failed to remove reaction: %v", err)
	}

	// Verify the reaction was removed
	comments, err = GetCommentsByPostID(db, postID, userID)
	if err != nil {
		t.Fatalf("Failed to get comments: %v", err)
	}
	comment = comments[0]
	if comment.Dislikes != 0 {
		t.Errorf("Expected 0 dislikes, got %d", comment.Dislikes)
	}
	if comment.UserReaction != 0 {
		t.Errorf("Expected user reaction 0, got %d", comment.UserReaction)
	}
}
