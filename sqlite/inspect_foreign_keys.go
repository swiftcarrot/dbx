package sqlite

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/swiftcarrot/dbx/schema"
)

// InspectForeignKeys retrieves all foreign keys for a table
func (s *SQLite) InspectForeignKeys(db *sql.DB, table *schema.Table) error {
	query := "PRAGMA foreign_key_list(" + table.Name + ")"

	rows, err := db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Use a map to group foreign key entries by id (SQLite returns one row per column)
	fkMap := make(map[int]*schema.ForeignKey)

	for rows.Next() {
		var id, seq int
		var tableName, from, to, onUpdate, onDelete string
		var match string // not used

		if err := rows.Scan(&id, &seq, &tableName, &from, &to, &onUpdate, &onDelete, &match); err != nil {
			return err
		}

		fk, exists := fkMap[id]
		if !exists {
			// Generate a constraint name since SQLite doesn't name constraints
			fkName := fmt.Sprintf("fk_%s_%s", table.Name, strings.ToLower(tableName))

			fk = &schema.ForeignKey{
				Name:       fkName,
				RefTable:   tableName,
				OnUpdate:   strings.ToUpper(onUpdate),
				OnDelete:   strings.ToUpper(onDelete),
				Columns:    []string{},
				RefColumns: []string{},
			}
			fkMap[id] = fk
		}

		fk.Columns = append(fk.Columns, from)
		fk.RefColumns = append(fk.RefColumns, to)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	// Add all foreign keys from the map to the table
	for _, fk := range fkMap {
		table.ForeignKeys = append(table.ForeignKeys, fk)
	}

	return nil
}
