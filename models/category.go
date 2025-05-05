package models

import (
	"database/sql"
)

type Category struct {
	ID   int64
	Name string
}

// CreateCategory creates a new category
func CreateCategory(db *sql.DB, name string) (int64, error) {
	result, err := db.Exec(
		"INSERT INTO categories (name) VALUES (?)",
		name,
	)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// GetAllCategories retrieves all categories
func GetAllCategories(db *sql.DB) ([]Category, error) {
	rows, err := db.Query("SELECT id, name FROM categories ORDER BY name ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var category Category
		if err := rows.Scan(&category.ID, &category.Name); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, nil
}

// GetCategoryByID retrieves a category by ID
func GetCategoryByID(db *sql.DB, id int64) (*Category, error) {
	var category Category
	err := db.QueryRow(
		"SELECT id, name FROM categories WHERE id = ?",
		id,
	).Scan(&category.ID, &category.Name)
	if err != nil {
		return nil, err
	}
	return &category, nil
}

// CategoryExists checks if a category with the given name exists
func CategoryExists(db *sql.DB, name string) (bool, error) {
	var exists bool
	err := db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM categories WHERE name = ?)",
		name,
	).Scan(&exists)
	return exists, err
}
