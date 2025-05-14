package sqlite

import (
	"database/sql"

	"github.com/swiftcarrot/dbx/schema"
)

// InspectViews retrieves all views from the database
func (s *SQLite) InspectViews(db *sql.DB, schm *schema.Schema) error {
	query := `
		SELECT name, sql FROM sqlite_master
		WHERE type = 'view'
		ORDER BY name
	`

	rows, err := db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var name, definition string
		if err := rows.Scan(&name, &definition); err != nil {
			return err
		}

		// Extract the view definition (everything after "AS ")
		definitionParts := extractViewDefinition(definition)

		// Get the view columns by executing a query against the view
		columnQuery := "PRAGMA table_info(" + name + ")"
		columnRows, err := db.Query(columnQuery)
		if err != nil {
			return err
		}

		var columns []string
		for columnRows.Next() {
			var cid int
			var columnName, dataType string
			var notNull, pk int
			var dflt interface{}

			if err := columnRows.Scan(&cid, &columnName, &dataType, &notNull, &dflt, &pk); err != nil {
				columnRows.Close()
				return err
			}

			columns = append(columns, columnName)
		}

		columnRows.Close()
		if err := columnRows.Err(); err != nil {
			return err
		}

		view := &schema.View{
			Name:       name,
			Definition: definitionParts,
			Columns:    columns,
		}

		schm.Views = append(schm.Views, view)
	}

	return rows.Err()
}

// extractViewDefinition extracts the portion of SQL after "AS" in a view definition
func extractViewDefinition(sql string) string {
	// Find the " AS " part of the view definition
	asIndex := -1
	for i := 0; i < len(sql)-3; i++ {
		if (sql[i] == ' ' || sql[i] == '\n' || sql[i] == '\t') &&
			(sql[i+1] == 'A' || sql[i+1] == 'a') &&
			(sql[i+2] == 'S' || sql[i+2] == 's') &&
			(sql[i+3] == ' ' || sql[i+3] == '\n' || sql[i+3] == '\t') {
			asIndex = i + 3
			break
		}
	}

	if asIndex != -1 {
		return sql[asIndex+1:] // +1 to skip the space after "AS"
	}
	return sql
}
