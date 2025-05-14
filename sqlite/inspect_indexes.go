package sqlite

import (
	"database/sql"
	"fmt"

	"github.com/swiftcarrot/dbx/schema"
)

// InspectIndexes retrieves all indexes for a table
func (s *SQLite) InspectIndexes(db *sql.DB, table *schema.Table) error {
	// First get the list of indexes
	query := "PRAGMA index_list(" + table.Name + ")"

	rows, err := db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var seq int
		var name string
		var unique bool
		var origin, partial string // origin and partial are not used directly

		if err := rows.Scan(&seq, &name, &unique, &origin, &partial); err != nil {
			return err
		}

		// Skip SQLite internal indexes which typically start with "sqlite_"
		if name == "sqlite_autoindex_"+table.Name+"_1" ||
			name == "sqlite_autoindex_"+table.Name+"_2" {
			// Skip primary key and unique constraint auto-indexes
			continue
		}

		// For each index, get its columns
		columnQuery := fmt.Sprintf("PRAGMA index_info(%s)", name)
		columnRows, err := db.Query(columnQuery)
		if err != nil {
			return err
		}

		var columns []string
		for columnRows.Next() {
			var seqno, cid int
			var colName string

			if err := columnRows.Scan(&seqno, &cid, &colName); err != nil {
				columnRows.Close()
				return err
			}

			columns = append(columns, colName)
		}

		columnRows.Close()
		if err := columnRows.Err(); err != nil {
			return err
		}

		// Only add the index if it has columns
		if len(columns) > 0 {
			index := &schema.Index{
				Name:    name,
				Columns: columns,
				Unique:  unique,
			}
			table.Indexes = append(table.Indexes, index)
		}
	}

	return rows.Err()
}
