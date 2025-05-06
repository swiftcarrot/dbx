package postgresql

import (
	"database/sql"
	"fmt"

	"github.com/swiftcarrot/dbx/schema"
)

// InspectSequences returns all sequences in the database
func (pg *PostgreSQL) InspectSequences(db *sql.DB, s *schema.Schema) error {
	// First check which column names are used in this PostgreSQL version
	var cacheColumn string
	checkCacheQuery := `
		SELECT column_name
		FROM information_schema.columns
		WHERE table_name = 'pg_sequences' AND column_name IN ('cache_value', 'cache_size')
		LIMIT 1
	`

	err := db.QueryRow(checkCacheQuery).Scan(&cacheColumn)
	if err != nil {
		// If we can't determine the column, default to cache_size which is more common
		cacheColumn = "cache_size"
	}

	// Check which column is used for cycle property
	var cycleColumn string
	checkCycleQuery := `
		SELECT column_name
		FROM information_schema.columns
		WHERE table_name = 'pg_sequences' AND column_name IN ('is_cycled', 'cycle')
		LIMIT 1
	`

	err = db.QueryRow(checkCycleQuery).Scan(&cycleColumn)
	if err != nil {
		// If we can't determine the column, default to is_cycled which is more common
		cycleColumn = "is_cycled"
	}

	query := fmt.Sprintf(`
		SELECT
			n.nspname AS sequence_schema,
			c.relname AS sequence_name,
			COALESCE(pg_catalog.obj_description(c.oid, 'pg_class'), '') AS description,
			s.start_value,
			s.increment_by,
			s.min_value,
			s.max_value,
			s.%s,
			s.%s
		FROM
			pg_class c
		JOIN
			pg_namespace n ON c.relnamespace = n.oid
		JOIN
			pg_sequences s ON s.schemaname = n.nspname AND s.sequencename = c.relname
		WHERE
			c.relkind = 'S' AND
			n.nspname != 'pg_catalog' AND
			n.nspname != 'information_schema'
		ORDER BY
			n.nspname, c.relname
	`, cacheColumn, cycleColumn)

	rows, err := db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var schemaName, seqName, description string
		var startVal, incrementBy, minVal, maxVal, cacheVal int64
		var isCycled bool

		if err := rows.Scan(
			&schemaName,
			&seqName,
			&description,
			&startVal,
			&incrementBy,
			&minVal,
			&maxVal,
			&cacheVal,
			&isCycled,
		); err != nil {
			return err
		}

		options := []schema.SequenceOption{
			schema.Start(startVal),
			schema.Increment(incrementBy),
			schema.MinValue(minVal),
			schema.MaxValue(maxVal),
			schema.Cache(cacheVal),
		}

		if isCycled {
			options = append(options, schema.Cycle)
		}

		if schemaName != "public" {
			options = append(options, schema.InSchema(schemaName))
		}

		s.CreateSequence(seqName, options...)
	}

	return rows.Err()
}
