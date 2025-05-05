package main

import (
    "log"
    "net/http"
)

func main() {
    // Initialize database
    db, err := initDB()
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Initialize routes
    mux := http.NewServeMux()
    
    // Serve static files
    mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

    // Auth routes
    mux.HandleFunc("/register", handleRegister)
    mux.HandleFunc("/login", handleLogin)
    mux.HandleFunc("/logout", handleLogout)

    // Post routes
    mux.HandleFunc("/", handleHome)
    mux.HandleFunc("/post/create", handleCreatePost)
    mux.HandleFunc("/post/", handleViewPost)
    mux.HandleFunc("/post/like", handleLikePost)
    mux.HandleFunc("/post/dislike", handleDislikePost)

    // Comment routes
    mux.HandleFunc("/comment/create", handleCreateComment)
    mux.HandleFunc("/comment/like", handleLikeComment)
    mux.HandleFunc("/comment/dislike", handleDislikeComment)

    log.Println("Server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", mux))
}