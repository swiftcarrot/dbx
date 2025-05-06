package postgresql

import "database/sql"

// InspectSchemas returns all schema names in the database
func (pg *PostgreSQL) InspectSchemas(db *sql.DB) ([]string, error) {
	query := `
		SELECT schema_name
		FROM information_schema.schemata
		WHERE schema_name NOT IN ('information_schema', 'pg_catalog', 'pg_toast', 'pg_temp_1', 'pg_toast_temp_1', 'pg_temp_2', 'pg_toast_temp_2')
		ORDER BY schema_name
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schemas []string
	for rows.Next() {
		var schemaName string
		if err := rows.Scan(&schemaName); err != nil {
			return nil, err
		}
		schemas = append(schemas, schemaName)
	}

	return schemas, rows.Err()
}
