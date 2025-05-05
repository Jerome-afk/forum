package handlers

import (
        "fmt"
        "net/http"
        "strconv"
        "strings"

        "forum/models"
)

// HomeHandler displays the home page with all posts
func HomeHandler(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path != "/" {
                http.NotFound(w, r)
                return
        }

        db := getDB(r)
        user := getUserFromContext(r)
        
        var userID int64
        if user != nil {
                userID = user.ID
        }

        // Get filter parameters
        categoryIDStr := r.URL.Query().Get("category")
        var categoryID int64
        if categoryIDStr != "" {
                var err error
                categoryID, err = strconv.ParseInt(categoryIDStr, 10, 64)
                if err != nil {
                        http.Error(w, "Invalid category ID", http.StatusBadRequest)
                        return
                }
        }

        // Get all posts with filter
        posts, err := models.GetPosts(db, userID, categoryID, 0, false)
        if err != nil {
                http.Error(w, fmt.Sprintf("Failed to get posts: %v", err), http.StatusInternalServerError)
                return
        }

        // Get all categories for filter dropdown
        categories, err := models.GetAllCategories(db)
        if err != nil {
                http.Error(w, fmt.Sprintf("Failed to get categories: %v", err), http.StatusInternalServerError)
                return
        }

        // Prepare data for template
        data := map[string]interface{}{
                "Posts":         posts,
                "Categories":    categories,
                "User":          user,
                "SelectedCategoryID": categoryID,
        }

        renderTemplate(w, "home.html", data)
}

// ViewPostHandler displays a single post with its comments
func ViewPostHandler(w http.ResponseWriter, r *http.Request) {
        // Extract post ID from URL path
        parts := strings.Split(r.URL.Path, "/")
        if len(parts) < 3 {
                http.NotFound(w, r)
                return
        }

        postIDStr := parts[2]
        postID, err := strconv.ParseInt(postIDStr, 10, 64)
        if err != nil {
                http.NotFound(w, r)
                return
        }

        db := getDB(r)
        user := getUserFromContext(r)
        
        var userID int64
        if user != nil {
                userID = user.ID
        }

        // Get the post with all details
        post, err := models.GetPostByID(db, postID, userID)
        if err != nil {
                // If post not found, return 404 instead of internal server error
                if err.Error() == "post not found" {
                        RenderErrorPage(w, http.StatusNotFound)
                } else {
                        http.Error(w, fmt.Sprintf("Failed to get post: %v", err), http.StatusInternalServerError)
                }
                return
        }

        // Get comments for the post
        comments, err := models.GetCommentsByPostID(db, postID, userID)
        if err != nil {
                http.Error(w, fmt.Sprintf("Failed to get comments: %v", err), http.StatusInternalServerError)
                return
        }

        // Check for error messages from the query parameters
        errorMsg := r.URL.Query().Get("error")
        
        // Prepare data for template
        data := map[string]interface{}{
                "Post":     post,
                "Comments": comments,
                "User":     user,
                "ErrorMsg": errorMsg,
                "ShowComments": strings.Contains(r.URL.Fragment, "comments"),
        }

        renderTemplate(w, "post.html", data)
}

// CreatePostHandler handles creation of new posts
func CreatePostHandler(w http.ResponseWriter, r *http.Request) {
        db := getDB(r)
        user := getUserFromContext(r)
        
        if r.Method == "GET" {
                // Get all categories for the form
                categories, err := models.GetAllCategories(db)
                if err != nil {
                        http.Error(w, fmt.Sprintf("Failed to get categories: %v", err), http.StatusInternalServerError)
                        return
                }

                data := map[string]interface{}{
                        "Categories": categories,
                        "User":       user,
                }

                renderTemplate(w, "create_post.html", data)
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
                title := strings.TrimSpace(r.FormValue("title"))
                content := strings.TrimSpace(r.FormValue("content"))
                categoryIDs := r.Form["categories"]

                // Basic validation
                var errors []string
                if title == "" {
                        errors = append(errors, "Title is required")
                }
                if content == "" {
                        errors = append(errors, "Content is required")
                }
                if len(categoryIDs) == 0 {
                        errors = append(errors, "At least one category is required")
                }

                if len(errors) > 0 {
                        // Get all categories for the form
                        categories, err := models.GetAllCategories(db)
                        if err != nil {
                                http.Error(w, fmt.Sprintf("Failed to get categories: %v", err), http.StatusInternalServerError)
                                return
                        }

                        data := map[string]interface{}{
                                "Errors":     errors,
                                "Title":      title,
                                "Content":    content,
                                "Categories": categories,
                                "User":       user,
                        }

                        renderTemplate(w, "create_post.html", data)
                        return
                }

                // Convert category IDs to int64
                var categoryIDsInt []int64
                for _, idStr := range categoryIDs {
                        id, err := strconv.ParseInt(idStr, 10, 64)
                        if err != nil {
                                http.Error(w, "Invalid category ID", http.StatusBadRequest)
                                return
                        }
                        categoryIDsInt = append(categoryIDsInt, id)
                }

                // Create the post
                postID, err := models.CreatePost(db, title, content, user.ID, categoryIDsInt)
                if err != nil {
                        http.Error(w, fmt.Sprintf("Failed to create post: %v", err), http.StatusInternalServerError)
                        return
                }

                // Redirect to the new post
                http.Redirect(w, r, fmt.Sprintf("/post/%d", postID), http.StatusSeeOther)
        }
}

// CategoryPostsHandler displays posts filtered by category
func CategoryPostsHandler(w http.ResponseWriter, r *http.Request) {
        // Extract category ID from URL path
        parts := strings.Split(r.URL.Path, "/")
        if len(parts) < 4 {
                http.NotFound(w, r)
                return
        }

        categoryIDStr := parts[3]
        categoryID, err := strconv.ParseInt(categoryIDStr, 10, 64)
        if err != nil {
                http.NotFound(w, r)
                return
        }

        db := getDB(r)
        user := getUserFromContext(r)
        
        var userID int64
        if user != nil {
                userID = user.ID
        }

        // Get the category
        category, err := models.GetCategoryByID(db, categoryID)
        if err != nil {
                http.Error(w, fmt.Sprintf("Failed to get category: %v", err), http.StatusInternalServerError)
                return
        }

        // Get posts filtered by category
        posts, err := models.GetPosts(db, userID, categoryID, 0, false)
        if err != nil {
                http.Error(w, fmt.Sprintf("Failed to get posts: %v", err), http.StatusInternalServerError)
                return
        }

        // Get all categories for filter dropdown
        categories, err := models.GetAllCategories(db)
        if err != nil {
                http.Error(w, fmt.Sprintf("Failed to get categories: %v", err), http.StatusInternalServerError)
                return
        }

        // Prepare data for template
        data := map[string]interface{}{
                "Posts":         posts,
                "Categories":    categories,
                "User":          user,
                "CurrentCategory": category,
                "SelectedCategoryID": categoryID,
        }

        renderTemplate(w, "home.html", data)
}

// MyPostsHandler displays posts created by the logged-in user
func MyPostsHandler(w http.ResponseWriter, r *http.Request) {
        db := getDB(r)
        user := getUserFromContext(r)

        // Get posts created by the current user
        posts, err := models.GetPosts(db, user.ID, 0, user.ID, false)
        if err != nil {
                http.Error(w, fmt.Sprintf("Failed to get posts: %v", err), http.StatusInternalServerError)
                return
        }

        // Get all categories for filter dropdown
        categories, err := models.GetAllCategories(db)
        if err != nil {
                http.Error(w, fmt.Sprintf("Failed to get categories: %v", err), http.StatusInternalServerError)
                return
        }

        // Prepare data for template
        data := map[string]interface{}{
                "Posts":      posts,
                "Categories": categories,
                "User":       user,
                "Title":      "My Posts",
        }

        renderTemplate(w, "home.html", data)
}

// LikedPostsHandler displays posts liked by the logged-in user
func LikedPostsHandler(w http.ResponseWriter, r *http.Request) {
        db := getDB(r)
        user := getUserFromContext(r)

        // Get posts liked by the current user
        posts, err := models.GetPosts(db, user.ID, 0, 0, true)
        if err != nil {
                http.Error(w, fmt.Sprintf("Failed to get posts: %v", err), http.StatusInternalServerError)
                return
        }

        // Get all categories for filter dropdown
        categories, err := models.GetAllCategories(db)
        if err != nil {
                http.Error(w, fmt.Sprintf("Failed to get categories: %v", err), http.StatusInternalServerError)
                return
        }

        // Prepare data for template
        data := map[string]interface{}{
                "Posts":      posts,
                "Categories": categories,
                "User":       user,
                "Title":      "Posts I Liked",
        }

        renderTemplate(w, "home.html", data)
}

// ReactPostHandler handles liking/disliking posts
func ReactPostHandler(w http.ResponseWriter, r *http.Request) {
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
        reactionStr := r.FormValue("reaction")
        
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
        err = models.ReactToPost(db, postID, user.ID, reaction)
        if err != nil {
                http.Error(w, fmt.Sprintf("Failed to save reaction: %v", err), http.StatusInternalServerError)
                return
        }

        // Redirect back to the post or referer
        redirectURL := fmt.Sprintf("/post/%d", postID)
        referer := r.Header.Get("Referer")
        if referer != "" {
                redirectURL = referer
        }
        
        http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}
