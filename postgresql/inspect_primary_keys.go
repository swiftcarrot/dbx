package postgresql

import (
	"database/sql"

	"github.com/swiftcarrot/dbx/schema"
)

// InspectPrimaryKey gets the primary key for a table
func (pg *PostgreSQL) InspectPrimaryKey(db *sql.DB, table *schema.Table) error {
	query := `
		SELECT
			tc.constraint_name,
			kcu.column_name
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu
			ON tc.constraint_name = kcu.constraint_name
		WHERE tc.table_schema = 'public'
		AND tc.table_name = $1
		AND tc.constraint_type = 'PRIMARY KEY'
		ORDER BY kcu.ordinal_position
	`

	rows, err := db.Query(query, table.Name)
	if err != nil {
		return err
	}
	defer rows.Close()

	var pkName string
	var columns []string

	for rows.Next() {
		var constraintName, columnName string
		if err := rows.Scan(&constraintName, &columnName); err != nil {
			return err
		}

		// Save constraint name only once
		if pkName == "" {
			pkName = constraintName
		}

		columns = append(columns, columnName)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	if len(columns) > 0 {
		table.SetPrimaryKey(pkName, columns)
	}

	return nil
}
