package testutil

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// GetSQLiteTestConn returns a connection to a temporary SQLite database for testing
func GetSQLiteTestConn() (*sql.DB, error) {
	// Create temporary directory if it doesn't exist
	tempDir := filepath.Join(os.TempDir(), "dbx_test")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, err
	}

	// Create a temporary SQLite database file
	dbPath := filepath.Join(tempDir, "test.db")

	// Open SQLite database connection
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Ensure the connection is working
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
