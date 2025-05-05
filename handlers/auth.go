package handlers

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"forum/models"
	"forum/utils"
)

// RegisterHandler handles user registration
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	db := getDB(r)

	if r.Method == "GET" {
		renderTemplate(w, "register.html", nil)
		return
	}

	if r.Method == "POST" {
		// Parse form
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		// Extract form values
		username := strings.TrimSpace(r.FormValue("username"))
		email := strings.TrimSpace(r.FormValue("email"))
		password := r.FormValue("password")
		confirmPassword := r.FormValue("confirm_password")

		// Basic validation
		var errors []string
		if username == "" {
			errors = append(errors, "Username is required")
		}
		if email == "" {
			errors = append(errors, "Email is required")
		}
		if password == "" {
			errors = append(errors, "Password is required")
		}
		if password != confirmPassword {
			errors = append(errors, "Passwords do not match")
		}

		if len(errors) > 0 {
			renderTemplate(w, "register.html", map[string]interface{}{
				"Errors":   errors,
				"Username": username,
				"Email":    email,
			})
			return
		}

		// Create user in database
		userID, err := models.CreateUser(db, username, email, password)
		if err != nil {
			renderTemplate(w, "register.html", map[string]interface{}{
				"Errors":   []string{err.Error()},
				"Username": username,
				"Email":    email,
			})
			return
		}

		// Create session for the new user
		sessionID, err := utils.CreateSession(db, userID)
		if err != nil {
			log.Printf("Failed to create session: %v", err)
			http.Error(w, "Failed to create session", http.StatusInternalServerError)
			return
		}

		// Set session cookie
		setSessionCookie(w, sessionID)

		// Redirect to home page
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

// LoginHandler handles user login
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	db := getDB(r)

	if r.Method == "GET" {
		renderTemplate(w, "login.html", nil)
		return
	}

	if r.Method == "POST" {
		// Parse form
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		// Extract form values
		email := strings.TrimSpace(r.FormValue("email"))
		password := r.FormValue("password")

		// Basic validation
		var errors []string
		if email == "" {
			errors = append(errors, "Email is required")
		}
		if password == "" {
			errors = append(errors, "Password is required")
		}

		if len(errors) > 0 {
			renderTemplate(w, "login.html", map[string]interface{}{
				"Errors": errors,
				"Email":  email,
			})
			return
		}

		// Authenticate user
		user, err := models.AuthenticateUser(db, email, password)
		if err != nil {
			renderTemplate(w, "login.html", map[string]interface{}{
				"Errors": []string{"Invalid email or password"},
				"Email":  email,
			})
			return
		}

		// Create session
		sessionID, err := utils.CreateSession(db, user.ID)
		if err != nil {
			log.Printf("Failed to create session: %v", err)
			http.Error(w, "Failed to create session", http.StatusInternalServerError)
			return
		}

		// Set session cookie
		setSessionCookie(w, sessionID)

		// Redirect to home page
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

// LogoutHandler handles user logout
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get session cookie
	cookie, err := r.Cookie("session_id")
	if err == nil {
		// Delete session from database
		db := getDB(r)
		utils.DeleteSession(db, cookie.Value)
	}

	// Clear session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	// Redirect to home page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Helper to set session cookie
func setSessionCookie(w http.ResponseWriter, sessionID string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		MaxAge:   int(time.Hour * 24 * 7 / time.Second), // 7 days
		HttpOnly: true,
	})
}

// Helper to get the database connection from request context
func getDB(r *http.Request) *sql.DB {
	return r.Context().Value(dbContextKey).(*sql.DB)
}

// Helper to render templates
func renderTemplate(w http.ResponseWriter, tmplFile string, data interface{}) {
	tmplPath := filepath.Join("templates", tmplFile)
	layoutPath := filepath.Join("templates", "layout.html")
	
	tmpl, err := template.ParseFiles(layoutPath, tmplPath)
	if err != nil {
		log.Printf("Failed to parse template: %v", err)
		http.Error(w, fmt.Sprintf("Failed to parse template: %v", err), http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteTemplate(w, "layout", data)
	if err != nil {
		log.Printf("Failed to execute template: %v", err)
		http.Error(w, fmt.Sprintf("Failed to execute template: %v", err), http.StatusInternalServerError)
	}
}
