package sqlite

import (
	"database/sql"
	"fmt"

	"github.com/swiftcarrot/dbx/schema"
)

// SQLite implements the Dialect interface for SQLite databases
type SQLite struct{}

// New creates a new SQLite dialect
func New() *SQLite {
	return &SQLite{}
}

// Inspect queries the SQLite database and returns its schema
func (s *SQLite) Inspect(db *sql.DB) (*schema.Schema, error) {
	// Create a new schema
	schema := schema.NewSchema()

	// Get tables
	tables, err := s.InspectTables(db)
	if err != nil {
		return nil, fmt.Errorf("failed to get tables: %w", err)
	}

	// For each table, get its columns, primary key, indexes, and foreign keys
	for _, tableName := range tables {
		table := schema.CreateTable(tableName, nil)

		// Get columns
		if err := s.InspectColumns(db, table); err != nil {
			return nil, fmt.Errorf("failed to get columns for table %s: %w", tableName, err)
		}

		// Get primary key
		if err := s.InspectPrimaryKey(db, table); err != nil {
			return nil, fmt.Errorf("failed to get primary key for table %s: %w", tableName, err)
		}

		// Get indexes
		if err := s.InspectIndexes(db, table); err != nil {
			return nil, fmt.Errorf("failed to get indexes for table %s: %w", tableName, err)
		}

		// Get foreign keys
		if err := s.InspectForeignKeys(db, table); err != nil {
			return nil, fmt.Errorf("failed to get foreign keys for table %s: %w", tableName, err)
		}
	}

	// Get views
	if err := s.InspectViews(db, schema); err != nil {
		return nil, fmt.Errorf("failed to get views: %w", err)
	}

	// Get triggers
	if err := s.InspectTriggers(db, schema); err != nil {
		return nil, fmt.Errorf("failed to get triggers: %w", err)
	}

	return schema, nil
}
