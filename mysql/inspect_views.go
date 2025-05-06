package mysql

import (
	"database/sql"
	"fmt"

	"github.com/swiftcarrot/dbx/schema"
)

// InspectViews inspects views in the database
func (my *MySQL) InspectViews(db *sql.DB, s *schema.Schema) error {
	// Query to get views and their definitions
	query := `
		SELECT
			table_name,
			view_definition
		FROM
			information_schema.views
		WHERE
			table_schema = DATABASE()
		ORDER BY
			table_name;
	`

	rows, err := db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			name       string
			definition sql.NullString
		)

		if err := rows.Scan(&name, &definition); err != nil {
			return err
		}

		// Skip if no definition
		if !definition.Valid {
			continue
		}

		// Create view
		view := &schema.View{
			Name:       name,
			Definition: definition.String,
		}

		// Get the column names for this view
		if err := my.inspectViewColumns(db, view); err != nil {
			return fmt.Errorf("error getting columns for view %s: %w", name, err)
		}

		// Add the view to the schema
		s.Views = append(s.Views, view)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating views: %w", err)
	}

	return nil
}

// inspectViewColumns gets the column names for a view
func (my *MySQL) inspectViewColumns(db *sql.DB, view *schema.View) error {
	query := `
		SELECT
			column_name
		FROM
			information_schema.columns
		WHERE
			table_name = ?
			AND table_schema = DATABASE()
		ORDER BY
			ordinal_position;
	`

	rows, err := db.Query(query, view.Name)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var columnName string

		if err := rows.Scan(&columnName); err != nil {
			return err
		}

		view.Columns = append(view.Columns, columnName)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating view columns: %w", err)
	}

	return nil
}
