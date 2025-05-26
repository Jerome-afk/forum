package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"forum/database"
	"forum/handlers"
	"forum/models"
	"forum/utils"
)

func main() {
	// Database initialization
	db, err := database.InitDB("./forum.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Run migrations
	err = database.RunMigrations(db)
	if err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize default categories if they don't exist
	defaultCategories := []string{"General", "Technology", "Sports", "Entertainment", "Science"}
	for _, category := range defaultCategories {
		exists, err := models.CategoryExists(db, category)
		if err != nil {
			log.Printf("Error checking if category exists: %v", err)
			continue
		}
		if !exists {
			_, err := models.CreateCategory(db, category)
			if err != nil {
				log.Printf("Error creating category %s: %v", category, err)
			}
		}
	}

	// Create directories for templates and static files if they don't exist
	dirs := []string{"./static", "./static/css", "./static/js", "./templates"}
	for _, dir := range dirs {
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			log.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Apply middleware to all handlers
	mux := http.NewServeMux()

	// Static file server
	fs := http.FileServer(http.Dir("./static"))

	// Wrap handler functions with middleware
	withMiddleware := func(handler http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Add database to context
			ctx := context.WithValue(r.Context(), handlers.GetDBContextKey(), db)
			r = r.WithContext(ctx)

			// Add session/user data to context
			sessionMiddleware := handlers.SessionMiddleware(db, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handler(w, r)
			}))
			sessionMiddleware.ServeHTTP(w, r)
		}
	}

	// Wrap static functions with middleware
	withStaticMiddleware := func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Add database to context
			ctx := context.WithValue(r.Context(), handlers.GetDBContextKey(), db)
			r = r.WithContext(ctx)

			// Add session/user data to context
			sessionMiddleware := handlers.SessionMiddleware(db, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handler.ServeHTTP(w, r)
			}))
			sessionMiddleware.ServeHTTP(w, r)
		})
	}

	// Static permissions
	protectedStatic := withStaticMiddleware(handlers.StaticMiddleware(fs))
	mux.Handle("/static/", http.StripPrefix("/static/", protectedStatic))

	// Auth routes
	mux.HandleFunc("/register", withMiddleware(handlers.RegisterHandler))
	mux.HandleFunc("/login", withMiddleware(handlers.LoginHandler))
	mux.HandleFunc("/logout", withMiddleware(handlers.LogoutHandler))

	// Post routes
	// mux.HandleFunc("/", withMiddleware(handlers.HomeHandler))
	mux.HandleFunc("/post/create", withMiddleware(handlers.AuthMiddleware(handlers.CreatePostHandler)))
	mux.HandleFunc("/post/", withMiddleware(handlers.ViewPostHandler))
	mux.HandleFunc("/posts/category/", withMiddleware(handlers.CategoryPostsHandler))
	mux.HandleFunc("/posts/my", withMiddleware(handlers.AuthMiddleware(handlers.MyPostsHandler)))
	mux.HandleFunc("/posts/liked", withMiddleware(handlers.AuthMiddleware(handlers.LikedPostsHandler)))

	// Comment routes
	mux.HandleFunc("/comment/create", withMiddleware(handlers.AuthMiddleware(handlers.CreateCommentHandler)))

	// Reaction routes (like/dislike)
	mux.HandleFunc("/post/react", withMiddleware(handlers.AuthMiddleware(handlers.ReactPostHandler)))
	mux.HandleFunc("/comment/react", withMiddleware(handlers.AuthMiddleware(handlers.ReactCommentHandler)))

	// Error pages
	mux.HandleFunc("/error", withMiddleware(handlers.NoPageHandler))
	mux.HandleFunc("/crashed", withMiddleware(handlers.ServerProblemHandler))

	// Add this after all your other routes
	mux.HandleFunc("/", withMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.Redirect(w, r, "/error", http.StatusFound)
			return
		}
		handlers.HomeHandler(w, r)
	}))

	// Clean expired sessions periodically
	go func() {
		for {
			utils.CleanExpiredSessions(db)
			time.Sleep(time.Hour) // Run every hour
		}
	}()

	// Start server
	port := "3000" // Changed to use port 5000
	server := &http.Server{
		Addr:         "0.0.0.0:" + port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	fmt.Printf("Starting server at http://0.0.0.0:%s\n", port)
	log.Fatal(server.ListenAndServe())
}
