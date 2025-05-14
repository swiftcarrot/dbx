package postgresql

import (
	"database/sql"

	"github.com/swiftcarrot/dbx/schema"
)

// InspectIndexes gets all indexes for a table (excluding primary key)
func (pg *PostgreSQL) InspectIndexes(db *sql.DB, table *schema.Table) error {
	query := `
		SELECT
            i.indexname AS name,
            array_agg(a.attname) AS columns,
            idx.indisunique AS is_unique
        FROM pg_indexes i
        JOIN pg_class c ON c.relname = i.tablename AND c.relnamespace = i.schemaname::regnamespace
        JOIN pg_index idx ON idx.indexrelid = (SELECT oid FROM pg_class WHERE relname = i.indexname AND relnamespace = i.schemaname::regnamespace)
        JOIN pg_attribute a ON a.attrelid = c.oid AND a.attnum = ANY(idx.indkey)
        WHERE
			i.schemaname NOT IN ('pg_catalog', 'information_schema')
			AND i.tablename = $1
			AND idx.indisprimary = false
        GROUP BY i.indexname, i.schemaname, idx.indisunique
		ORDER BY i.indexname;
	`

	rows, err := db.Query(query, table.Name)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var columns string
		var isUnique bool

		if err := rows.Scan(&name, &columns, &isUnique); err != nil {
			return err
		}

		options := []schema.IndexOption{}
		if isUnique {
			options = append(options, schema.Unique)
		}
		table.Index(name, PostgresArrayToSlice(columns), options...)
	}

	return rows.Err()
}
