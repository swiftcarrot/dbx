package mysql

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/swiftcarrot/dbx/schema"
)

// InspectColumns inspects columns for a table
func (my *MySQL) InspectColumns(db *sql.DB, table *schema.Table) error {
	query := `
		SELECT
			column_name,
			data_type,
			is_nullable,
			column_default,
			character_maximum_length,
			numeric_precision,
			numeric_scale,
			column_type,
			extra
		FROM
			information_schema.columns
		WHERE
			table_schema = DATABASE()
			AND table_name = ?
		ORDER BY
			ordinal_position;
	`

	rows, err := db.Query(query, table.Name)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			name             string
			dataType         string
			isNullable       string
			defaultValue     sql.NullString
			charMaxLength    sql.NullInt64
			numericPrecision sql.NullInt64
			numericScale     sql.NullInt64
			columnType       string
			extra            string
		)

		if err := rows.Scan(
			&name,
			&dataType,
			&isNullable,
			&defaultValue,
			&charMaxLength,
			&numericPrecision,
			&numericScale,
			&columnType,
			&extra,
		); err != nil {
			return err
		}

		// Create a column
		column := &schema.Column{
			Name: name,
		}

		// Set type, using columnType which includes length/precision info
		// rather than just dataType
		column.Type = columnType

		// Set nullable
		column.Nullable = isNullable == "YES"

		// Handle default value
		if defaultValue.Valid {
			value := defaultValue.String

			// Handle special cases like CURRENT_TIMESTAMP
			if strings.ToUpper(value) == "CURRENT_TIMESTAMP" ||
				strings.HasPrefix(strings.ToUpper(value), "CURRENT_TIMESTAMP(") {
				column.Default = "CURRENT_TIMESTAMP"
			} else {
				column.Default = value
			}
		}

		// Handle auto-increment
		if strings.Contains(extra, "auto_increment") {
			column.Identity = "BY DEFAULT"
		}

		// Add the column to the table
		table.Columns = append(table.Columns, column)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating columns: %w", err)
	}

	return nil
}
