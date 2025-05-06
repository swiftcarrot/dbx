package mysql

import (
	"database/sql"
	"fmt"

	"github.com/swiftcarrot/dbx/schema"
)

// MySQL represents the MySQL dialect
type MySQL struct{}

// New creates a new MySQL inspector
func New() *MySQL {
	return &MySQL{}
}

// Inspect inspects the database and returns a schema
func (my *MySQL) Inspect(db *sql.DB) (*schema.Schema, error) {
	s := schema.NewSchema()

	// Get tables
	tables, err := my.InspectTables(db)
	if err != nil {
		return nil, fmt.Errorf("error inspecting tables: %w", err)
	}
	for _, tableName := range tables {
		table := s.CreateTable(tableName, nil)

		// Get columns
		if err := my.InspectColumns(db, table); err != nil {
			return nil, fmt.Errorf("failed to get columns for table %s: %w", tableName, err)
		}

		// Get primary key
		if err := my.InspectPrimaryKey(db, table); err != nil {
			return nil, fmt.Errorf("failed to get primary key for table %s: %w", tableName, err)
		}

		// Get indexes
		if err := my.InspectIndexes(db, table); err != nil {
			return nil, fmt.Errorf("failed to get indexes for table %s: %w", tableName, err)
		}

		// Get foreign keys
		if err := my.InspectForeignKeys(db, table); err != nil {
			return nil, fmt.Errorf("failed to get foreign keys for table %s: %w", tableName, err)
		}
	}

	// Get views
	if err := my.InspectViews(db, s); err != nil {
		return nil, fmt.Errorf("error inspecting views: %w", err)
	}

	// Get functions
	if err := my.InspectFunctions(db, s); err != nil {
		return nil, fmt.Errorf("error inspecting functions: %w", err)
	}

	if err := my.InspectTriggers(db, s); err != nil {
		return nil, fmt.Errorf("error inspecting triggers: %w", err)
	}

	return s, nil
}
