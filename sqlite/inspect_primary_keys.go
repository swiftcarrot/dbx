package sqlite

import (
	"database/sql"

	"github.com/swiftcarrot/dbx/schema"
)

// InspectPrimaryKey retrieves the primary key for a table
func (s *SQLite) InspectPrimaryKey(db *sql.DB, table *schema.Table) error {
	query := "PRAGMA table_info(" + table.Name + ")"

	rows, err := db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	var pkColumns []string
	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull, pk int
		var dflt interface{}

		if err := rows.Scan(&cid, &name, &dataType, &notNull, &dflt, &pk); err != nil {
			return err
		}

		if pk > 0 {
			pkColumns = append(pkColumns, name)
		}
	}

	if err := rows.Err(); err != nil {
		return err
	}

	// Only create primary key if we found primary key columns
	if len(pkColumns) > 0 {
		table.PrimaryKey = &schema.PrimaryKey{
			Name:    table.Name + "_pkey", // SQLite doesn't name constraints, but we create a name for consistency
			Columns: pkColumns,
		}
	}

	return nil
}
