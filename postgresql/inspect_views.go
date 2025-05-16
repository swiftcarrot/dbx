package postgresql

import (
	"database/sql"
	"fmt"

	"github.com/swiftcarrot/dbx/schema"
)

// InspectViews retrieves all views from the database
func (pg *PostgreSQL) InspectViews(db *sql.DB, s *schema.Schema) error {
	query := `
    SELECT
        n.nspname AS schema,
        c.relname AS name,
        pg_get_viewdef(c.oid, true) AS definition,
        array_agg(a.attname ORDER BY a.attnum) AS columns
    FROM
        pg_class c
        JOIN pg_namespace n ON c.relnamespace = n.oid
        LEFT JOIN pg_attribute a ON c.oid = a.attrelid AND a.attnum > 0 AND NOT a.attisdropped
    WHERE
        c.relkind = 'v'
        AND n.nspname NOT IN ('pg_catalog', 'information_schema')
    GROUP BY
        n.nspname, c.relname, c.oid
    ORDER BY
        n.nspname, c.relname;
    `

	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("query failed: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var v schema.View
		var columns sql.NullString
		if err := rows.Scan(&v.Schema, &v.Name, &v.Definition, &columns); err != nil {
			return fmt.Errorf("scan failed: %v", err)
		}

		// Handle columns array (PostgreSQL returns as string like "{col1,col2}")
		if columns.Valid {
			// Remove curly braces and split by comma
			colStr := columns.String
			if len(colStr) > 2 {
				colStr = colStr[1 : len(colStr)-1] // Remove { and }
				if colStr != "" {
					v.Columns = splitColumns(colStr)
				}
			}
		}

		// Options are empty for now (extend if needed)
		v.Options = []string{}

		s.Views = append(s.Views, &v)
	}

	return rows.Err()
}

// splitColumns splits a PostgreSQL array string into a slice, handling commas correctly
func splitColumns(s string) []string {
	var result []string
	var current string
	inQuote := false
	for _, r := range s {
		switch r {
		case '"':
			inQuote = !inQuote
			current += string(r)
		case ',':
			if inQuote {
				current += string(r)
			} else {
				result = append(result, current)
				current = ""
			}
		default:
			current += string(r)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}
