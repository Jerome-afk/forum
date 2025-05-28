package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"forum/database"
	"forum/models"
)

// setupAuthTestDB creates a temporary database for testing auth handlers
func setupAuthTestDB(t *testing.T) (*sql.DB, func()) {
	// Create a temporary database file for testing
	tempDBPath := "./test_auth.db"

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

// createRequestWithDB creates an HTTP request with the database in context
func createRequestWithDB(method, url string, body *bytes.Buffer, db *sql.DB) *http.Request {
	var req *http.Request
	if body != nil {
		req, _ = http.NewRequest(method, url, body)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else {
		req, _ = http.NewRequest(method, url, nil)
	}

	// Add database to context
	ctx := context.WithValue(req.Context(), dbContextKey, db)
	return req.WithContext(ctx)
}

func TestRegisterHandler(t *testing.T) {
	db, cleanup := setupAuthTestDB(t)
	defer cleanup()

	// Test GET request to register page
	req := createRequestWithDB("GET", "/register", nil, db)
	rr := httptest.NewRecorder()

	RegisterHandler(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Test successful registration
	formData := url.Values{}
	formData.Set("username", "testuser")
	formData.Set("email", "test@example.com")
	formData.Set("password", "password123")
	formData.Set("confirm_password", "password123")

	req = createRequestWithDB("POST", "/register", bytes.NewBufferString(formData.Encode()), db)
	rr = httptest.NewRecorder()

	RegisterHandler(rr, req)

	if status := rr.Code; status != http.StatusSeeOther {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusSeeOther)
	}

	// Verify the user was created in the database
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE username = ? AND email = ?", "testuser", "test@example.com").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query user: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 user, got %d", count)
	}

	// Test registration with duplicate username
	formData = url.Values{}
	formData.Set("username", "testuser")
	formData.Set("email", "another@example.com")
	formData.Set("password", "password123")
	formData.Set("confirm_password", "password123")

	req = createRequestWithDB("POST", "/register", bytes.NewBufferString(formData.Encode()), db)
	rr = httptest.NewRecorder()

	RegisterHandler(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	if !strings.Contains(rr.Body.String(), "username already taken") {
		t.Errorf("Expected error message about duplicate username, got: %s", rr.Body.String())
	}

	// Test registration with duplicate email
	formData = url.Values{}
	formData.Set("username", "anotheruser")
	formData.Set("email", "test@example.com")
	formData.Set("password", "password123")
	formData.Set("confirm_password", "password123")

	req = createRequestWithDB("POST", "/register", bytes.NewBufferString(formData.Encode()), db)
	rr = httptest.NewRecorder()

	RegisterHandler(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	if !strings.Contains(rr.Body.String(), "email already registered") {
		t.Errorf("Expected error message about duplicate email, got: %s", rr.Body.String())
	}

	// Test registration with mismatched passwords
	formData = url.Values{}
	formData.Set("username", "newuser")
	formData.Set("email", "new@example.com")
	formData.Set("password", "password123")
	formData.Set("confirm_password", "differentpassword")

	req = createRequestWithDB("POST", "/register", bytes.NewBufferString(formData.Encode()), db)
	rr = httptest.NewRecorder()

	RegisterHandler(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	if !strings.Contains(rr.Body.String(), "Passwords do not match") {
		t.Errorf("Expected error message about mismatched passwords, got: %s", rr.Body.String())
	}
}

func TestLoginHandler(t *testing.T) {
	db, cleanup := setupAuthTestDB(t)
	defer cleanup()

	// Create a test user
	_, err := models.CreateUser(db, "testuser", "test@example.com", "password123")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Test GET request to login page
	req := createRequestWithDB("GET", "/login", nil, db)
	rr := httptest.NewRecorder()

	LoginHandler(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Test successful login
	formData := url.Values{}
	formData.Set("email", "test@example.com")
	formData.Set("password", "password123")

	req = createRequestWithDB("POST", "/login", bytes.NewBufferString(formData.Encode()), db)
	rr = httptest.NewRecorder()

	LoginHandler(rr, req)

	if status := rr.Code; status != http.StatusSeeOther {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusSeeOther)
	}

	// Verify a session was created
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM sessions").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query sessions: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 session, got %d", count)
	}

	// Test login with wrong password
	formData = url.Values{}
	formData.Set("email", "test@example.com")
	formData.Set("password", "wrongpassword")

	req = createRequestWithDB("POST", "/login", bytes.NewBufferString(formData.Encode()), db)
	rr = httptest.NewRecorder()

	LoginHandler(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	if !strings.Contains(rr.Body.String(), "Invalid email or password") {
		t.Errorf("Expected error message about invalid credentials, got: %s", rr.Body.String())
	}

	// Test login with non-existent email
	formData = url.Values{}
	formData.Set("email", "nonexistent@example.com")
	formData.Set("password", "password123")

	req = createRequestWithDB("POST", "/login", bytes.NewBufferString(formData.Encode()), db)
	rr = httptest.NewRecorder()

	LoginHandler(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	if !strings.Contains(rr.Body.String(), "Invalid email or password") {
		t.Errorf("Expected error message about invalid credentials, got: %s", rr.Body.String())
	}
}

func TestLogoutHandler(t *testing.T) {
	db, cleanup := setupAuthTestDB(t)
	defer cleanup()

	// Create a test user and session
	userID, err := models.CreateUser(db, "testuser", "test@example.com", "password123")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	_, err = db.Exec(
		"INSERT INTO sessions (id, user_id, expires_at) VALUES (?, ?, ?)",
		"test-session-id", userID, time.Now().Add(time.Hour*24),
	)
	if err != nil {
		t.Fatalf("Failed to create test session: %v", err)
	}

	// Create a request with a session cookie
	req := createRequestWithDB("POST", "/logout", nil, db)
	cookie := &http.Cookie{
		Name:  "session_id",
		Value: "test-session-id",
	}
	req.AddCookie(cookie)

	rr := httptest.NewRecorder()

	LogoutHandler(rr, req)

	if status := rr.Code; status != http.StatusSeeOther {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusSeeOther)
	}

	// Verify the session was deleted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM sessions WHERE id = ?", "test-session-id").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query session: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected session to be deleted, but it still exists")
	}

	// Verify the session cookie was cleared
	for _, cookie := range rr.Result().Cookies() {
		if cookie.Name == "session_id" {
			if cookie.Value != "" || cookie.MaxAge >= 0 {
				t.Errorf("Expected session cookie to be cleared, got value: %s, maxAge: %d", cookie.Value, cookie.MaxAge)
			}
		}
	}
}
