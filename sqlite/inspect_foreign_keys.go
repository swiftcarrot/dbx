package sqlite

import (
	"database/sql"

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

	for rows.Next() {
		var id, seq int
		var tableName, from, to, onUpdate, onDelete string
		var match string // not used

		if err := rows.Scan(&id, &seq, &tableName, &from, &to, &onUpdate, &onDelete, &match); err != nil {
			return err
		}

		fk := &schema.ForeignKey{
			RefTable:   tableName,
			OnUpdate:   onUpdate,
			OnDelete:   onDelete,
			Columns:    []string{},
			RefColumns: []string{},
		}

		fk.Columns = append(fk.Columns, from)
		fk.RefColumns = append(fk.RefColumns, to)

		table.ForeignKeys = append(table.ForeignKeys, fk)
	}

	return rows.Err()
}
