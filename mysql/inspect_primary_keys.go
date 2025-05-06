package mysql

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/swiftcarrot/dbx/schema"
)

// InspectPrimaryKey inspects the primary key for a table
func (my *MySQL) InspectPrimaryKey(db *sql.DB, table *schema.Table) error {
	query := `
		SELECT
			kc.constraint_name,
			GROUP_CONCAT(kc.column_name ORDER BY kc.ordinal_position) AS columns
		FROM
			information_schema.table_constraints tc
		JOIN
			information_schema.key_column_usage kc
			ON tc.constraint_name = kc.constraint_name
			AND tc.constraint_type = 'PRIMARY KEY'
			AND tc.table_schema = kc.table_schema
		WHERE
			tc.table_name = ?
			AND tc.table_schema = DATABASE()
		GROUP BY
			kc.constraint_name;
	`

	var name string
	var columnsStr string

	err := db.QueryRow(query, table.Name).Scan(&name, &columnsStr)
	if err != nil {
		if err == sql.ErrNoRows {
			// No primary key for this table
			return nil
		}
		return fmt.Errorf("error getting primary key: %w", err)
	}

	columns := splitAndTrim(columnsStr, ",")
	if len(columns) > 0 {
		table.SetPrimaryKey(name, columns)
	}

	return nil
}

// Helper function to split a comma-separated string and trim whitespace
func splitAndTrim(s, sep string) []string {
	parts := make([]string, 0)
	for _, part := range strings.Split(s, sep) {
		parts = append(parts, strings.TrimSpace(part))
	}
	return parts
}
