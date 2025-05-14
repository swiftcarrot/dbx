package sqlite

import (
	"database/sql"
	"fmt"
	"strconv"
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
		var typeStr string // Use string to scan the type initially

		if err := rows.Scan(&cid, &col.Name, &typeStr, &notNull, &dfltValue, &pk); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		col.Nullable = notNull == 0
		if dfltValue.Valid {
			col.Default = dfltValue.String
		}

		// Parse the SQLite type string
		parsedType, limit, typePrecision, typeScale, err := parseSQLiteType(typeStr)
		if err != nil {
			return fmt.Errorf("failed to parse SQLite type for %s: %w", col.Name, err)
		}

		// Set precision and scale for numeric types
		col.Precision = typePrecision
		col.Scale = typeScale

		// Create the appropriate schema.ColumnType
		switch strings.ToLower(parsedType) {
		case "text":
			col.Type = &TextType{}
		case "integer", "int":
			col.Type = &IntegerType{}
		case "bigint":
			col.Type = &schema.BigIntType{}
		case "smallint":
			col.Type = &schema.SmallIntType{}
		case "real", "float":
			col.Type = &schema.FloatType{}
		case "numeric", "decimal":
			col.Type = &schema.DecimalType{
				Precision: typePrecision,
				Scale:     typeScale,
			}
		case "varchar", "character varying":
			col.Type = &schema.VarcharType{
				Length: limit,
			}
		case "boolean", "bool":
			col.Type = &schema.BooleanType{}
		case "date":
			col.Type = &schema.DateType{}
		case "time":
			col.Type = &schema.TimeType{}
		case "timestamp":
			col.Type = &schema.TimestampType{}
		default:
			// Use TextType as fallback
			col.Type = &schema.TextType{}
		}

		// Create the appropriate schema.ColumnType
		switch strings.ToLower(typeStr) {
		case "text":
			col.Type = &TextType{}
		case "integer", "int":
			col.Type = &IntegerType{}
		case "real":
			col.Type = &schema.FloatType{}
		case "numeric", "decimal":
			col.Type = &schema.DecimalType{
				Precision: typePrecision,
				Scale:     typeScale,
			}
		case "varchar", "character varying":
			col.Type = &schema.VarcharType{
				Length: limit,
			}
		case "boolean", "bool":
			col.Type = &schema.BooleanType{}
		case "date":
			col.Type = &schema.DateType{}
		case "time":
			col.Type = &schema.TimeType{}
		case "timestamp":
			col.Type = &schema.TimestampType{}
		default:
			// Use TextType as fallback
			col.Type = &schema.TextType{}
		}

		// Set precision and scale for numeric types
		if typePrecision > 0 {
			col.Precision = typePrecision
		}
		if typeScale > 0 {
			col.Scale = typeScale
		}

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
func parseSQLiteType(sqliteType string) (typeName string, limit, precision, scale int, err error) {
	// Convert to uppercase and remove extra spaces
	sqliteType = strings.ToUpper(strings.TrimSpace(sqliteType))

	// Extract type and parameters (e.g., "VARCHAR(255)" or "DECIMAL(10,2)")
	parenIdx := strings.Index(sqliteType, "(")
	if parenIdx == -1 {
		return sqliteType, 0, 0, 0, nil
	}

	typeName = sqliteType[:parenIdx]
	params := strings.Trim(sqliteType[parenIdx+1:len(sqliteType)-1], " ")

	if typeName == "DECIMAL" || typeName == "NUMERIC" {
		// Handle DECIMAL(precision, scale)
		parts := strings.Split(params, ",")
		if len(parts) == 2 {
			precision, err = strconv.Atoi(strings.TrimSpace(parts[0]))
			if err != nil {
				return typeName, 0, 0, 0, err
			}
			scale, err = strconv.Atoi(strings.TrimSpace(parts[1]))
			if err != nil {
				return typeName, 0, 0, 0, err
			}
		} else if len(parts) == 1 {
			precision, err = strconv.Atoi(strings.TrimSpace(parts[0]))
			if err != nil {
				return typeName, 0, 0, 0, err
			}
		}
	} else if typeName == "VARCHAR" || typeName == "CHAR" || typeName == "CHARACTER" {
		// Handle VARCHAR(length)
		limit, err = strconv.Atoi(params)
		if err != nil {
			return typeName, 0, 0, 0, err
		}
	}

	return typeName, limit, precision, scale, nil
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
