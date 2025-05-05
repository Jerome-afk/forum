package database

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

// InitDB initializes the SQLite database connection
func InitDB(filepath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		return nil, err
	}

	// Check if database is accessible
	if err = db.Ping(); err != nil {
		return nil, err
	}

	log.Println("Connected to database")
	return db, nil
}
