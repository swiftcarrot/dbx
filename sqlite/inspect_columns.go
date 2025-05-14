package sqlite

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/swiftcarrot/dbx/schema"
)

// InspectColumns retrieves all columns for a table
func (s *SQLite) InspectColumns(db *sql.DB, table *schema.Table) error {
	rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(%s)", table.Name))
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var col schema.Column
		var cid int
		var notNull int
		var dfltValue sql.NullString
		var pk int

		if err := rows.Scan(&cid, &col.Name, &col.Type, &notNull, &dfltValue, &pk); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		col.Nullable = notNull == 0
		if dfltValue.Valid {
			col.Default = dfltValue.String
		}

		col.Type, col.Limit, col.Precision, col.Scale = parseSQLiteType(col.Type)

		col.AutoIncrement, err = isAutoIncrement(db, table.Name, col.Name)
		if err != nil {
			return fmt.Errorf("failed to check AUTOINCREMENT for %s: %w", col.Name, err)
		}

		col.Comment = ""

		table.Columns = append(table.Columns, &col)
	}

	return rows.Err()
}

// parseSQLiteType extracts type, limit, precision, and scale from SQLite type string
func parseSQLiteType(sqliteType string) (typeName string, limit, precision, scale int) {
	// Convert to uppercase and remove extra spaces
	sqliteType = strings.ToUpper(strings.TrimSpace(sqliteType))

	// Extract type and parameters (e.g., "VARCHAR(255)" or "DECIMAL(10,2)")
	parenIdx := strings.Index(sqliteType, "(")
	if parenIdx == -1 {
		return sqliteType, 0, 0, 0
	}

	typeName = sqliteType[:parenIdx]
	params := strings.Trim(sqliteType[parenIdx+1:len(sqliteType)-1], " ")

	if typeName == "DECIMAL" || typeName == "NUMERIC" {
		// Handle DECIMAL(precision, scale)
		parts := strings.Split(params, ",")
		if len(parts) == 2 {
			fmt.Sscanf(parts[0], "%d", &precision)
			fmt.Sscanf(parts[1], "%d", &scale)
		} else if len(parts) == 1 {
			fmt.Sscanf(parts[0], "%d", &precision)
		}
	} else if typeName == "VARCHAR" || typeName == "CHAR" || typeName == "CHARACTER" {
		// Handle VARCHAR(length)
		fmt.Sscanf(params, "%d", &limit)
	}

	return typeName, limit, precision, scale
}

// isAutoIncrement checks if a column is defined with AUTOINCREMENT
func isAutoIncrement(db *sql.DB, tableName, columnName string) (bool, error) {
	// Query the table's SQL creation statement
	query := "SELECT sql FROM sqlite_master WHERE type='table' AND name=?"
	var sqlStmt string
	err := db.QueryRow(query, tableName).Scan(&sqlStmt)
	if err != nil {
		return false, fmt.Errorf("failed to get table SQL: %w", err)
	}

	// Check for AUTOINCREMENT in the column definition
	// Normalize the SQL statement to handle case sensitivity and spaces
	sqlStmt = strings.ToUpper(sqlStmt)
	pattern := fmt.Sprintf(`%s\s+[^,]*\s+AUTOINCREMENT`, strings.ToUpper(columnName))
	return strings.Contains(sqlStmt, pattern), nil
}
