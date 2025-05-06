package postgresql

import (
	"database/sql"

	"github.com/swiftcarrot/dbx/schema"
)

// InspectExtensions returns all installed PostgreSQL extensions
func (pg *PostgreSQL) InspectExtensions(db *sql.DB, s *schema.Schema) error {
	query := `
		SELECT extname
		FROM pg_extension
		ORDER BY extname
	`

	rows, err := db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var extName string
		if err := rows.Scan(&extName); err != nil {
			return err
		}
		s.EnableExtension(extName)
	}

	return rows.Err()
}
