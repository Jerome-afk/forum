package handlers

import (
	"context"
	"database/sql"
	"net/http"

	"forum/models"
	"forum/utils"
)

// Context keys
type contextKey string
const (
        dbContextKey   contextKey = "db"
        userContextKey contextKey = "user"
)

// GetDBContextKey returns the database context key
func GetDBContextKey() contextKey {
        return dbContextKey
}

// GetUserContextKey returns the user context key
func GetUserContextKey() contextKey {
        return userContextKey
}

// DBMiddleware injects the database into the request context
func DBMiddleware(db *sql.DB, next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                ctx := context.WithValue(r.Context(), dbContextKey, db)
                next.ServeHTTP(w, r.WithContext(ctx))
        })
}

// SessionMiddleware checks for a valid session and adds the user to context
func SessionMiddleware(db *sql.DB, next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                // Get session cookie
                cookie, err := r.Cookie("session_id")
                if err != nil {
                        // No session cookie, continue without user
                        next.ServeHTTP(w, r)
                        return
                }

                // Validate session
                userID, err := utils.ValidateSession(db, cookie.Value)
                if err != nil {
                        // Invalid session, clear cookie and continue without user
                        http.SetCookie(w, &http.Cookie{
                                Name:     "session_id",
                                Value:    "",
                                Path:     "/",
                                MaxAge:   -1,
                                HttpOnly: true,
                        })
                        next.ServeHTTP(w, r)
                        return
                }

                // Get user information
                user, err := models.GetUserByID(db, userID)
                if err != nil {
                        // User not found, clear cookie and continue without user
                        http.SetCookie(w, &http.Cookie{
                                Name:     "session_id",
                                Value:    "",
                                Path:     "/",
                                MaxAge:   -1,
                                HttpOnly: true,
                        })
                        next.ServeHTTP(w, r)
                        return
                }

                // Add user to context and continue
                ctx := context.WithValue(r.Context(), userContextKey, user)
                next.ServeHTTP(w, r.WithContext(ctx))
        })
}

// AuthMiddleware protects routes that require authentication
func AuthMiddleware(handler http.HandlerFunc) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
                user := getUserFromContext(r)
                if user == nil {
                        http.Redirect(w, r, "/login", http.StatusSeeOther)
                        return
                }
                handler(w, r)
        }
}

// Helper function to get user from context
func getUserFromContext(r *http.Request) *models.User {
        user, ok := r.Context().Value(userContextKey).(*models.User)
        if !ok {
                return nil
        }
        return user
}
