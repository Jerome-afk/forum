package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"forum/models"
)

// CreateCommentHandler handles creating new comments on posts
func CreateCommentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Extract form values
	postIDStr := r.FormValue("post_id")
	content := strings.TrimSpace(r.FormValue("content"))
	
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}
	
	// Basic validation
	if content == "" {
		// Redirect back to the post with an error message
		http.Redirect(w, r, fmt.Sprintf("/post/%d?error=Comment cannot be empty", postID), http.StatusSeeOther)
		return
	}

	db := getDB(r)
	user := getUserFromContext(r)

	// Create the comment
	_, err = models.CreateComment(db, content, user.ID, postID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create comment: %v", err), http.StatusInternalServerError)
		return
	}

	// Redirect back to the post
	http.Redirect(w, r, fmt.Sprintf("/post/%d", postID), http.StatusSeeOther)
}

// ReactCommentHandler handles liking/disliking comments
func ReactCommentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Extract form values
	commentIDStr := r.FormValue("comment_id")
	postIDStr := r.FormValue("post_id")
	reactionStr := r.FormValue("reaction")
	
	commentID, err := strconv.ParseInt(commentIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}
	
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}
	
	reaction, err := strconv.Atoi(reactionStr)
	if err != nil || (reaction != 1 && reaction != -1) {
		http.Error(w, "Invalid reaction", http.StatusBadRequest)
		return
	}

	db := getDB(r)
	user := getUserFromContext(r)

	// Save the reaction
	err = models.ReactToComment(db, commentID, user.ID, reaction)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to save reaction: %v", err), http.StatusInternalServerError)
		return
	}

	// Redirect back to the post
	http.Redirect(w, r, fmt.Sprintf("/post/%d", postID), http.StatusSeeOther)
}
