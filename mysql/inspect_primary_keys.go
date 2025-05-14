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
			index_name,
			GROUP_CONCAT(column_name ORDER BY seq_in_index) as column_names
		FROM
			information_schema.statistics
		WHERE
			table_schema = DATABASE()
			AND table_name = ?
			AND index_name = 'PRIMARY'
		GROUP BY table_name
		ORDER BY table_name
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
