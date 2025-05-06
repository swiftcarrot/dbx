package postgresql

import (
	"database/sql"

	"github.com/swiftcarrot/dbx/schema"
)

// InspectForeignKeys gets all foreign keys for a table
func (pg *PostgreSQL) InspectForeignKeys(db *sql.DB, table *schema.Table) error {
	query := `
		SELECT
			tc.constraint_name,
			kcu.column_name,
			ccu.table_name AS foreign_table_name,
			ccu.column_name AS foreign_column_name,
			rc.update_rule,
			rc.delete_rule
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu
			ON tc.constraint_name = kcu.constraint_name
		JOIN information_schema.constraint_column_usage ccu
			ON tc.constraint_name = ccu.constraint_name
		JOIN information_schema.referential_constraints rc
			ON tc.constraint_name = rc.constraint_name
		WHERE tc.table_schema = 'public'
		AND tc.table_name = $1
		AND tc.constraint_type = 'FOREIGN KEY'
		ORDER BY tc.constraint_name, kcu.ordinal_position
	`

	rows, err := db.Query(query, table.Name)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Group by constraint name
	fkMap := make(map[string]struct {
		columns    []string
		refTable   string
		refColumns []string
		onUpdate   string
		onDelete   string
	})

	for rows.Next() {
		var (
			constraintName string
			columnName     string
			refTableName   string
			refColumnName  string
			updateRule     string
			deleteRule     string
		)

		if err := rows.Scan(&constraintName, &columnName, &refTableName, &refColumnName, &updateRule, &deleteRule); err != nil {
			return err
		}

		fk, exists := fkMap[constraintName]
		if !exists {
			fk = struct {
				columns    []string
				refTable   string
				refColumns []string
				onUpdate   string
				onDelete   string
			}{
				refTable: refTableName,
				onUpdate: updateRule,
				onDelete: deleteRule,
			}
		}

		fk.columns = append(fk.columns, columnName)
		fk.refColumns = append(fk.refColumns, refColumnName)
		fkMap[constraintName] = fk
	}

	if err := rows.Err(); err != nil {
		return err
	}

	// Create foreign keys
	for fkName, fk := range fkMap {
		options := []schema.ForeignKeyOption{}

		if fk.onDelete != "NO ACTION" {
			options = append(options, schema.OnDelete(fk.onDelete))
		}

		if fk.onUpdate != "NO ACTION" {
			options = append(options, schema.OnUpdate(fk.onUpdate))
		}

		table.ForeignKey(fkName, fk.columns, fk.refTable, fk.refColumns, options...)
	}

	return nil
}
