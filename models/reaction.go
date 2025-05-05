package models

import (
	"database/sql"
)

// GetPostReactionStats returns the number of likes and dislikes for a post
func GetPostReactionStats(db *sql.DB, postID int64) (likes int, dislikes int, err error) {
	err = db.QueryRow(`
		SELECT 
			(SELECT COUNT(*) FROM post_reactions WHERE post_id = ? AND reaction = 1),
			(SELECT COUNT(*) FROM post_reactions WHERE post_id = ? AND reaction = -1)
	`, postID, postID).Scan(&likes, &dislikes)
	return
}

// GetCommentReactionStats returns the number of likes and dislikes for a comment
func GetCommentReactionStats(db *sql.DB, commentID int64) (likes int, dislikes int, err error) {
	err = db.QueryRow(`
		SELECT 
			(SELECT COUNT(*) FROM comment_reactions WHERE comment_id = ? AND reaction = 1),
			(SELECT COUNT(*) FROM comment_reactions WHERE comment_id = ? AND reaction = -1)
	`, commentID, commentID).Scan(&likes, &dislikes)
	return
}

// GetUserPostReaction returns the user's reaction to a post (1, -1, or 0 if none)
func GetUserPostReaction(db *sql.DB, postID, userID int64) (int, error) {
	var reaction int
	err := db.QueryRow(
		"SELECT COALESCE(reaction, 0) FROM post_reactions WHERE post_id = ? AND user_id = ?",
		postID, userID,
	).Scan(&reaction)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return reaction, err
}

// GetUserCommentReaction returns the user's reaction to a comment (1, -1, or 0 if none)
func GetUserCommentReaction(db *sql.DB, commentID, userID int64) (int, error) {
	var reaction int
	err := db.QueryRow(
		"SELECT COALESCE(reaction, 0) FROM comment_reactions WHERE comment_id = ? AND user_id = ?",
		commentID, userID,
	).Scan(&reaction)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return reaction, err
}
