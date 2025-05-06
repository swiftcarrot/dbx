package mysql

import (
	"database/sql"
	"fmt"
)

// InspectTables returns a list of all tables in the database
func (my *MySQL) InspectTables(db *sql.DB) ([]string, error) {
	query := `
		SELECT
			table_name
		FROM
			information_schema.tables
		WHERE
			table_schema = DATABASE()
			AND table_type = 'BASE TABLE'
		ORDER BY
			table_name;
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string

	for rows.Next() {
		var name string

		if err := rows.Scan(&name); err != nil {
			return nil, err
		}

		tables = append(tables, name)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tables: %w", err)
	}

	return tables, nil
}
