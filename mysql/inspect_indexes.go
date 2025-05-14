package mysql

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/swiftcarrot/dbx/schema"
)

// InspectIndexes inspects indexes for a table
func (my *MySQL) InspectIndexes(db *sql.DB, table *schema.Table) error {
	query := `
		SELECT
			i.index_name,
			i.non_unique,
			GROUP_CONCAT(i.column_name ORDER BY i.seq_in_index) as column_names
		FROM
			information_schema.statistics i
		WHERE
			i.table_schema = DATABASE()
			AND i.table_name = ?
			AND i.index_name != 'PRIMARY'  -- Skip primary key
		GROUP BY
			i.index_name, i.non_unique
		ORDER BY
			i.index_name;
	`

	rows, err := db.Query(query, table.Name)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Map to track indexes we've already seen
	indexMap := make(map[string]bool)

	for rows.Next() {
		var (
			name      string
			nonUnique int
			columnStr string
		)

		if err := rows.Scan(&name, &nonUnique, &columnStr); err != nil {
			return err
		}

		// MySQL creates indexes for foreign keys with names like table_ibfk_1
		// Skip these indexes as they'll be handled by the foreign key inspection
		if strings.HasPrefix(name, fmt.Sprintf("%s_ibfk_", table.Name)) {
			continue
		}

		// Skip if we've already handled this index
		if _, exists := indexMap[name]; exists {
			continue
		}

		indexMap[name] = true

		// Get all columns for this index
		columns := splitAndTrim(columnStr, ",")

		// Create the index
		idx := &schema.Index{
			Name:    name,
			Columns: columns,
			Unique:  nonUnique == 0,
		}

		// Add it to the table
		table.Indexes = append(table.Indexes, idx)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating indexes: %w", err)
	}

	return nil
}
