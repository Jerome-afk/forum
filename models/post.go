package models

import (
	"database/sql"
	"time"
)

type Post struct {
	ID         int64
	Title      string
	Content    string
	UserID     int64
	Username   string
	CreatedAt  time.Time
	Categories []Category
	Likes      int
	Dislikes   int
	UserReaction int // 1 for like, -1 for dislike, 0 for none
}

// CreatePost creates a new post in the database
func CreatePost(db *sql.DB, title, content string, userID int64, categoryIDs []int64) (int64, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// Insert the post
	result, err := tx.Exec(
		"INSERT INTO posts (title, content, user_id) VALUES (?, ?, ?)",
		title, content, userID,
	)
	if err != nil {
		return 0, err
	}

	postID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	// Associate categories with the post
	for _, categoryID := range categoryIDs {
		_, err = tx.Exec(
			"INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)",
			postID, categoryID,
		)
		if err != nil {
			return 0, err
		}
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}

	return postID, nil
}

// GetPostByID retrieves a post by ID with categories, likes, and dislikes
func GetPostByID(db *sql.DB, postID int64, currentUserID int64) (*Post, error) {
	// Get post details
	var post Post
	err := db.QueryRow(`
		SELECT p.id, p.title, p.content, p.user_id, u.username, p.created_at,
		(SELECT COUNT(*) FROM post_reactions WHERE post_id = p.id AND reaction = 1) as likes,
		(SELECT COUNT(*) FROM post_reactions WHERE post_id = p.id AND reaction = -1) as dislikes,
		COALESCE((SELECT reaction FROM post_reactions WHERE post_id = p.id AND user_id = ?), 0) as user_reaction
		FROM posts p
		JOIN users u ON p.user_id = u.id
		WHERE p.id = ?
	`, currentUserID, postID).Scan(
		&post.ID, &post.Title, &post.Content, &post.UserID, &post.Username, 
		&post.CreatedAt, &post.Likes, &post.Dislikes, &post.UserReaction,
	)
	if err != nil {
		return nil, err
	}

	// Get categories for the post
	rows, err := db.Query(`
		SELECT c.id, c.name
		FROM categories c
		JOIN post_categories pc ON c.id = pc.category_id
		WHERE pc.post_id = ?
	`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	post.Categories = []Category{}
	for rows.Next() {
		var category Category
		if err := rows.Scan(&category.ID, &category.Name); err != nil {
			return nil, err
		}
		post.Categories = append(post.Categories, category)
	}

	return &post, nil
}

// GetPosts retrieves all posts with optional filtering
func GetPosts(db *sql.DB, currentUserID int64, categoryID int64, userID int64, likedOnly bool) ([]Post, error) {
	query := `
		SELECT p.id, p.title, p.content, p.user_id, u.username, p.created_at,
		(SELECT COUNT(*) FROM post_reactions WHERE post_id = p.id AND reaction = 1) as likes,
		(SELECT COUNT(*) FROM post_reactions WHERE post_id = p.id AND reaction = -1) as dislikes,
		COALESCE((SELECT reaction FROM post_reactions WHERE post_id = p.id AND user_id = ?), 0) as user_reaction
		FROM posts p
		JOIN users u ON p.user_id = u.id
	`

	args := []interface{}{currentUserID}
	whereClause := ""

	if categoryID > 0 {
		whereClause += " WHERE p.id IN (SELECT post_id FROM post_categories WHERE category_id = ?)"
		args = append(args, categoryID)
	}

	if userID > 0 {
		if whereClause == "" {
			whereClause = " WHERE p.user_id = ?"
		} else {
			whereClause += " AND p.user_id = ?"
		}
		args = append(args, userID)
	}

	if likedOnly {
		if whereClause == "" {
			whereClause = " WHERE p.id IN (SELECT post_id FROM post_reactions WHERE user_id = ? AND reaction = 1)"
		} else {
			whereClause += " AND p.id IN (SELECT post_id FROM post_reactions WHERE user_id = ? AND reaction = 1)"
		}
		args = append(args, currentUserID)
	}

	query += whereClause + " ORDER BY p.created_at DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		if err := rows.Scan(
			&post.ID, &post.Title, &post.Content, &post.UserID, &post.Username, 
			&post.CreatedAt, &post.Likes, &post.Dislikes, &post.UserReaction,
		); err != nil {
			return nil, err
		}

		// Get categories for each post
		catRows, err := db.Query(`
			SELECT c.id, c.name
			FROM categories c
			JOIN post_categories pc ON c.id = pc.category_id
			WHERE pc.post_id = ?
		`, post.ID)
		if err != nil {
			return nil, err
		}

		post.Categories = []Category{}
		for catRows.Next() {
			var category Category
			if err := catRows.Scan(&category.ID, &category.Name); err != nil {
				catRows.Close()
				return nil, err
			}
			post.Categories = append(post.Categories, category)
		}
		catRows.Close()

		posts = append(posts, post)
	}

	return posts, nil
}

// ReactToPost allows a user to like or dislike a post
func ReactToPost(db *sql.DB, postID, userID int64, reaction int) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Check if reaction already exists
	var exists bool
	var currentReaction int
	err = tx.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM post_reactions WHERE post_id = ? AND user_id = ?), COALESCE((SELECT reaction FROM post_reactions WHERE post_id = ? AND user_id = ?), 0)",
		postID, userID, postID, userID,
	).Scan(&exists, &currentReaction)
	if err != nil {
		return err
	}

	if exists {
		if currentReaction == reaction {
			// Remove reaction if clicking the same button
			_, err = tx.Exec(
				"DELETE FROM post_reactions WHERE post_id = ? AND user_id = ?",
				postID, userID,
			)
		} else {
			// Update existing reaction
			_, err = tx.Exec(
				"UPDATE post_reactions SET reaction = ? WHERE post_id = ? AND user_id = ?",
				reaction, postID, userID,
			)
		}
	} else {
		// Insert new reaction
		_, err = tx.Exec(
			"INSERT INTO post_reactions (post_id, user_id, reaction) VALUES (?, ?, ?)",
			postID, userID, reaction,
		)
	}
	if err != nil {
		return err
	}

	return tx.Commit()
}
