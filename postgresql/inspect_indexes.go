package postgresql

import (
	"database/sql"

	"github.com/swiftcarrot/dbx/schema"
)

// InspectIndexes gets all indexes for a table (excluding primary key)
func (pg *PostgreSQL) InspectIndexes(db *sql.DB, table *schema.Table) error {
	// This query excludes primary key constraints by checking constraint_type
	query := `
		SELECT
			i.relname AS index_name,
			a.attname AS column_name,
			ix.indisunique AS is_unique
		FROM
			pg_index ix
			JOIN pg_class i ON i.oid = ix.indexrelid
			JOIN pg_class t ON t.oid = ix.indrelid
			JOIN pg_attribute a ON a.attrelid = t.oid AND a.attnum = ANY(ix.indkey)
			JOIN pg_namespace n ON t.relnamespace = n.oid
			LEFT JOIN pg_constraint c ON c.conindid = ix.indexrelid
		WHERE
			t.relname = $1
			AND n.nspname = 'public'
			AND (c.contype IS NULL OR c.contype != 'p')
		ORDER BY
			i.relname, a.attnum
	`

	rows, err := db.Query(query, table.Name)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Group columns by index name
	indexColumns := make(map[string][]string)
	indexUnique := make(map[string]bool)

	for rows.Next() {
		var indexName, colName string
		var isUnique bool

		if err := rows.Scan(&indexName, &colName, &isUnique); err != nil {
			return err
		}

		indexColumns[indexName] = append(indexColumns[indexName], colName)
		indexUnique[indexName] = isUnique
	}

	if err := rows.Err(); err != nil {
		return err
	}

	// Create indexes
	for indexName, columns := range indexColumns {
		options := []schema.IndexOption{}
		if indexUnique[indexName] {
			options = append(options, schema.Unique)
		}
		table.Index(indexName, columns, options...)
	}

	return nil
}
