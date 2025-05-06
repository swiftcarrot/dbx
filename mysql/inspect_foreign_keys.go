package mysql

import (
	"database/sql"
	"fmt"

	"github.com/swiftcarrot/dbx/schema"
)

// InspectForeignKeys inspects foreign keys for a table
func (my *MySQL) InspectForeignKeys(db *sql.DB, table *schema.Table) error {
	query := `
		SELECT
			tc.constraint_name,
			GROUP_CONCAT(kcu.column_name ORDER BY kcu.ordinal_position) AS columns,
			kcu.referenced_table_name,
			GROUP_CONCAT(kcu.referenced_column_name ORDER BY kcu.ordinal_position) AS referenced_columns,
			rc.update_rule,
			rc.delete_rule
		FROM
			information_schema.table_constraints tc
		JOIN
			information_schema.key_column_usage kcu
			ON tc.constraint_name = kcu.constraint_name
			AND tc.constraint_type = 'FOREIGN KEY'
		JOIN
			information_schema.referential_constraints rc
			ON tc.constraint_name = rc.constraint_name
		WHERE
			tc.table_name = ?
			AND tc.table_schema = DATABASE()
		GROUP BY
			tc.constraint_name, kcu.referenced_table_name, rc.update_rule, rc.delete_rule
		ORDER BY
			tc.constraint_name;
	`

	rows, err := db.Query(query, table.Name)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			name          string
			columnsStr    string
			refTable      string
			refColumnsStr string
			updateRule    string
			deleteRule    string
		)

		if err := rows.Scan(&name, &columnsStr, &refTable, &refColumnsStr, &updateRule, &deleteRule); err != nil {
			return err
		}

		// Parse column lists
		columns := splitAndTrim(columnsStr, ",")
		refColumns := splitAndTrim(refColumnsStr, ",")

		// Create a foreign key
		fk := &schema.ForeignKey{
			Name:       name,
			Columns:    columns,
			RefTable:   refTable,
			RefColumns: refColumns,
		}

		// Set ON DELETE action if specified
		if deleteRule != "RESTRICT" {
			fk.OnDelete = deleteRule
		}

		// Set ON UPDATE action if specified
		if updateRule != "RESTRICT" {
			fk.OnUpdate = updateRule
		}

		// Add the foreign key to the table
		table.ForeignKeys = append(table.ForeignKeys, fk)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating foreign keys: %w", err)
	}

	return nil
}
