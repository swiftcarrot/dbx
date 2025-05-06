package postgresql

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// RowData represents a row of data with column name to value mapping
type RowData map[string]interface{}

// InspectRows retrieves rows from a table in the database
// Limit is required to avoid retrieving too many rows
func (pg *PostgreSQL) InspectRows(db *sql.DB, tableName string, limit int, whereClause string, args ...interface{}) ([]RowData, error) {
	// Build the query
	query := fmt.Sprintf("SELECT * FROM %s", quoteIdentifier(tableName))

	// Add WHERE clause if provided
	if whereClause != "" {
		query += " WHERE " + whereClause
	}

	// Add limit
	query += fmt.Sprintf(" LIMIT %d", limit)

	// Execute the query
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	// Prepare result
	var result []RowData

	// Scan rows
	for rows.Next() {
		// Create a slice of interface{} to hold the values
		values := make([]interface{}, len(columns))
		valuePointers := make([]interface{}, len(columns))

		// Set the pointers to the values
		for i := range values {
			valuePointers[i] = &values[i]
		}

		// Scan the row into the value pointers
		if err := rows.Scan(valuePointers...); err != nil {
			return nil, err
		}

		// Create a map for this row
		rowData := make(RowData)

		// Convert the values to appropriate types
		for i, col := range columns {
			val := valuePointers[i].(*interface{})
			rowData[col] = pg.convertValue(*val)
		}

		result = append(result, rowData)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// convertValue handles conversion of PostgreSQL data types to Go types
func (pg *PostgreSQL) convertValue(val interface{}) interface{} {
	switch v := val.(type) {
	case []byte:
		// Try to convert to string first
		str := string(v)

		// Check if it's a JSON object or array
		if (len(str) > 0 && (str[0] == '{' && str[len(str)-1] == '}')) ||
			(len(str) > 0 && (str[0] == '[' && str[len(str)-1] == ']')) {
			var jsonData interface{}
			if err := json.Unmarshal(v, &jsonData); err == nil {
				return jsonData
			}
		}

		// Try to convert to int
		if i, err := strconv.ParseInt(str, 10, 64); err == nil {
			return i
		}

		// Try to convert to float
		if f, err := strconv.ParseFloat(str, 64); err == nil {
			return f
		}

		// Try to convert to bool
		if str == "t" || str == "true" || str == "yes" || str == "y" || str == "1" {
			return true
		}
		if str == "f" || str == "false" || str == "no" || str == "n" || str == "0" {
			return false
		}

		// Return as string if no other conversion applies
		return str

	case time.Time:
		return v

	default:
		return v
	}
}
