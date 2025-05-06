package postgresql

import (
	"database/sql"
	"strings"

	"github.com/swiftcarrot/dbx/schema"
)

// InspectViews retrieves all views from the database
func (pg *PostgreSQL) InspectViews(db *sql.DB, s *schema.Schema) error {
	query := `
		SELECT
			n.nspname AS schema_name,
			c.relname AS view_name,
			pg_get_viewdef(c.oid) AS definition,
			array_to_string(array_agg(a.attname), ',') AS column_names,
			COALESCE(array_to_string(c.reloptions, ', '), '') AS options
		FROM
			pg_class c
		JOIN
			pg_namespace n ON c.relnamespace = n.oid
		LEFT JOIN
			pg_attribute a ON a.attrelid = c.oid AND a.attnum > 0 AND NOT a.attisdropped
		WHERE
			c.relkind = 'v' AND
			n.nspname NOT IN ('pg_catalog', 'information_schema')
		GROUP BY
			n.nspname, c.relname, c.oid, c.reloptions
		ORDER BY
			n.nspname, c.relname
	`

	rows, err := db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var schemaName, viewName, definition, columnNamesStr, optionsStr string

		if err := rows.Scan(&schemaName, &viewName, &definition, &columnNamesStr, &optionsStr); err != nil {
			return err
		}

		var options []string
		if optionsStr != "" {
			options = strings.Split(optionsStr, ", ")
		}

		var columnNames []string
		if columnNamesStr != "" {
			columnNames = strings.Split(columnNamesStr, ",")
		}

		s.CreateView(
			viewName,
			definition,
			schema.ViewInSchema(schemaName),
			schema.ViewColumns(columnNames...),
			schema.ViewOptions(options...),
		)
	}

	return rows.Err()
}
