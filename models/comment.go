package models

import (
	"database/sql"
	"time"
)

type Comment struct {
	ID           int64
	Content      string
	UserID       int64
	Username     string
	PostID       int64
	CreatedAt    time.Time
	Likes        int
	Dislikes     int
	UserReaction int // 1 for like, -1 for dislike, 0 for none
}

// CreateComment creates a new comment on a post
func CreateComment(db *sql.DB, content string, userID, postID int64) (int64, error) {
	result, err := db.Exec(
		"INSERT INTO comments (content, user_id, post_id) VALUES (?, ?, ?)",
		content, userID, postID,
	)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// GetCommentsByPostID retrieves all comments for a post
func GetCommentsByPostID(db *sql.DB, postID, currentUserID int64) ([]Comment, error) {
	rows, err := db.Query(`
		SELECT c.id, c.content, c.user_id, u.username, c.post_id, c.created_at,
		(SELECT COUNT(*) FROM comment_reactions WHERE comment_id = c.id AND reaction = 1) as likes,
		(SELECT COUNT(*) FROM comment_reactions WHERE comment_id = c.id AND reaction = -1) as dislikes,
		COALESCE((SELECT reaction FROM comment_reactions WHERE comment_id = c.id AND user_id = ?), 0) as user_reaction
		FROM comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.post_id = ?
		ORDER BY c.created_at ASC
	`, currentUserID, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var comment Comment
		if err := rows.Scan(
			&comment.ID, &comment.Content, &comment.UserID, &comment.Username, 
			&comment.PostID, &comment.CreatedAt, &comment.Likes, &comment.Dislikes, &comment.UserReaction,
		); err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}

	return comments, nil
}

// ReactToComment allows a user to like or dislike a comment
func ReactToComment(db *sql.DB, commentID, userID int64, reaction int) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Check if reaction already exists
	var exists bool
	var currentReaction int
	err = tx.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM comment_reactions WHERE comment_id = ? AND user_id = ?), COALESCE((SELECT reaction FROM comment_reactions WHERE comment_id = ? AND user_id = ?), 0)",
		commentID, userID, commentID, userID,
	).Scan(&exists, &currentReaction)
	if err != nil {
		return err
	}

	if exists {
		if currentReaction == reaction {
			// Remove reaction if clicking the same button
			_, err = tx.Exec(
				"DELETE FROM comment_reactions WHERE comment_id = ? AND user_id = ?",
				commentID, userID,
			)
		} else {
			// Update existing reaction
			_, err = tx.Exec(
				"UPDATE comment_reactions SET reaction = ? WHERE comment_id = ? AND user_id = ?",
				reaction, commentID, userID,
			)
		}
	} else {
		// Insert new reaction
		_, err = tx.Exec(
			"INSERT INTO comment_reactions (comment_id, user_id, reaction) VALUES (?, ?, ?)",
			commentID, userID, reaction,
		)
	}
	if err != nil {
		return err
	}

	return tx.Commit()
}
