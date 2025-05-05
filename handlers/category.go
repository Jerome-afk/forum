package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"forum/models"
)

// CreateCategoryHandler handles creating new categories (admin only)
func CreateCategoryHandler(w http.ResponseWriter, r *http.Request) {
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
	name := strings.TrimSpace(r.FormValue("name"))
	
	// Basic validation
	if name == "" {
		http.Error(w, "Category name is required", http.StatusBadRequest)
		return
	}

	db := getDB(r)

	// Check if category already exists
	exists, err := models.CategoryExists(db, name)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to check if category exists: %v", err), http.StatusInternalServerError)
		return
	}
	if exists {
		http.Error(w, "Category already exists", http.StatusBadRequest)
		return
	}

	// Create the category
	_, err = models.CreateCategory(db, name)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create category: %v", err), http.StatusInternalServerError)
		return
	}

	// Redirect back to home page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
