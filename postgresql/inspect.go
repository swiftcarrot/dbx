package postgresql

import (
	"database/sql"
	"fmt"

	"github.com/swiftcarrot/dbx/schema"
)

// PostgreSQL implements the Dialect interface for PostgreSQL databases
type PostgreSQL struct{}

// New creates a new PostgreSQL dialect
func New() *PostgreSQL {
	return &PostgreSQL{}
}

// Inspect queries the PostgreSQL database and returns its schema
func (pg *PostgreSQL) Inspect(db *sql.DB) (*schema.Schema, error) {
	// Create a new schema with default "public" schema name
	s := schema.NewSchema()

	// Get installed extensions
	if err := pg.InspectExtensions(db, s); err != nil {
		return nil, fmt.Errorf("failed to get extensions: %w", err)
	}

	// Get sequences
	if err := pg.InspectSequences(db, s); err != nil {
		return nil, fmt.Errorf("failed to get sequences: %w", err)
	}

	// Get functions
	if err := pg.InspectFunctions(db, s); err != nil {
		return nil, fmt.Errorf("failed to get functions: %w", err)
	}

	// Get views
	if err := pg.InspectViews(db, s); err != nil {
		return nil, fmt.Errorf("failed to get views: %w", err)
	}

	// Get row policies
	if err := pg.InspectRowPolicies(db, s); err != nil {
		return nil, fmt.Errorf("failed to get row policies: %w", err)
	}

	// Get tables in public schema
	tables, err := pg.InspectTables(db)
	if err != nil {
		return nil, fmt.Errorf("failed to get tables: %w", err)
	}

	// For each table, get its columns, primary key, indexes, and foreign keys
	for _, tableName := range tables {
		table := s.CreateTable(tableName, nil)

		// Get columns
		if err := pg.InspectColumns(db, table); err != nil {
			return nil, fmt.Errorf("failed to get columns for table %s: %w", tableName, err)
		}

		// Get primary key
		if err := pg.InspectPrimaryKey(db, table); err != nil {
			return nil, fmt.Errorf("failed to get primary key for table %s: %w", tableName, err)
		}

		// Get indexes
		if err := pg.InspectIndexes(db, table); err != nil {
			return nil, fmt.Errorf("failed to get indexes for table %s: %w", tableName, err)
		}

		// Get foreign keys
		if err := pg.InspectForeignKeys(db, table); err != nil {
			return nil, fmt.Errorf("failed to get foreign keys for table %s: %w", tableName, err)
		}
	}

	// Get triggers (after tables to ensure proper dependencies)
	if err := pg.InspectTriggers(db, s); err != nil {
		return nil, fmt.Errorf("failed to get triggers: %w", err)
	}

	return s, nil
}
