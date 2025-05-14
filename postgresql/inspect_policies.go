package postgresql

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/swiftcarrot/dbx/schema"
)

// InspectRowPolicies retrieves all row policies from the database
func (pg *PostgreSQL) InspectRowPolicies(db *sql.DB, s *schema.Schema) error {
	query := `
        SELECT
            schemaname,
            tablename,
            policyname,
            cmd,
            roles,
            qual,
            with_check,
            CASE WHEN permissive = 'PERMISSIVE' THEN true ELSE false END AS permissive
        FROM pg_policies
        ORDER BY schemaname, tablename, policyname
    `

	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("failed to query pg_policies: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var p schema.RowPolicy
		var rolesStr string
		var usingExpr, checkExpr sql.NullString

		err := rows.Scan(
			&p.Schema,
			&p.TableName,
			&p.PolicyName,
			&p.CommandType,
			&rolesStr,
			&usingExpr,
			&checkExpr,
			&p.Permissive,
		)
		if err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		// Parse roles (stored as a comma-separated string in braces, e.g., "{role1,role2}")
		rolesStr = strings.Trim(rolesStr, "{}")
		if rolesStr != "" {
			p.Roles = strings.Split(rolesStr, ",")
		} else {
			p.Roles = []string{}
		}

		// Handle nullable expressions
		p.UsingExpr = usingExpr.String
		if !usingExpr.Valid {
			p.UsingExpr = ""
		}
		p.CheckExpr = checkExpr.String
		if !checkExpr.Valid {
			p.CheckExpr = ""
		}

		s.RowPolicies = append(s.RowPolicies, &p)
	}

	return rows.Err()
}
