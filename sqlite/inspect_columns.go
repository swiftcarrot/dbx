package sqlite

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/swiftcarrot/dbx/schema"
)

// InspectColumns retrieves all columns for a table
func (s *SQLite) InspectColumns(db *sql.DB, table *schema.Table) error {
	query := "PRAGMA table_info(" + table.Name + ")"

	rows, err := db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull, pk int
		var dflt interface{}

		if err := rows.Scan(&cid, &name, &dataType, &notNull, &dflt, &pk); err != nil {
			return err
		}

		column := &schema.Column{
			Name:     name,
			Type:     dataType,
			Nullable: notNull == 0,
		}

		// Handle default value
		if dflt != nil {
			defaultValue, ok := dflt.(string)
			if ok {
				column.Default = defaultValue
			}
		}

		// Get column comment if available (SQLite doesn't natively support column comments)

		// Parse numeric precision and scale from type definition
		// SQLite doesn't have explicit precision/scale in schema, but we can parse it from type name
		if strings.HasPrefix(strings.ToLower(dataType), "numeric(") ||
			strings.HasPrefix(strings.ToLower(dataType), "decimal(") {
			parts := strings.Split(strings.TrimSuffix(strings.TrimPrefix(dataType, "numeric("), ")"), ",")
			if len(parts) >= 1 {
				fmt.Sscanf(parts[0], "%d", &column.Precision)
				if len(parts) >= 2 {
					fmt.Sscanf(parts[1], "%d", &column.Scale)
				}
			}
		}

		table.Columns = append(table.Columns, column)
	}

	return rows.Err()
}
