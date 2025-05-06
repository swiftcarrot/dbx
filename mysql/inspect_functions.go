package mysql

import (
	"database/sql"
	"fmt"

	"github.com/swiftcarrot/dbx/schema"
)

// InspectFunctions inspects stored functions in the database
func (my *MySQL) InspectFunctions(db *sql.DB, s *schema.Schema) error {
	// Query to get stored functions
	query := `
		SELECT
			routine_name,
			data_type,
			routine_definition,
			is_deterministic
		FROM
			information_schema.routines
		WHERE
			routine_schema = DATABASE()
			AND routine_type = 'FUNCTION'
		ORDER BY
			routine_name;
	`

	rows, err := db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			name            string
			returnType      string
			body            sql.NullString
			isDeterministic string
		)

		if err := rows.Scan(&name, &returnType, &body, &isDeterministic); err != nil {
			return err
		}

		// Skip if no body
		if !body.Valid {
			continue
		}

		// Create the function
		function := &schema.Function{
			Name:    name,
			Returns: returnType,
			Body:    body.String,
		}

		// Set volatility based on deterministic flag
		if isDeterministic == "YES" {
			function.Volatility = "IMMUTABLE"
		} else {
			function.Volatility = "STABLE"
		}

		// Get function parameters
		if err := my.inspectFunctionParameters(db, function); err != nil {
			return fmt.Errorf("error getting parameters for function %s: %w", name, err)
		}

		// Add the function to the schema
		s.Functions = append(s.Functions, function)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating functions: %w", err)
	}

	return nil
}

// inspectFunctionParameters gets the parameters for a function
func (my *MySQL) inspectFunctionParameters(db *sql.DB, function *schema.Function) error {
	query := `
		SELECT
			parameter_name,
			data_type,
			character_maximum_length,
			numeric_precision,
			parameter_mode
		FROM
			information_schema.parameters
		WHERE
			specific_name = ?
			AND specific_schema = DATABASE()
			AND ordinal_position > 0  -- Skip the return parameter
		ORDER BY
			ordinal_position;
	`

	rows, err := db.Query(query, function.Name)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			name          sql.NullString
			dataType      string
			charMaxLength sql.NullInt64
			numPrecision  sql.NullInt64
			parameterMode sql.NullString
		)

		if err := rows.Scan(&name, &dataType, &charMaxLength, &numPrecision, &parameterMode); err != nil {
			return err
		}

		// Format the full data type with precision/length if applicable
		fullType := dataType
		if charMaxLength.Valid && charMaxLength.Int64 > 0 {
			fullType = fmt.Sprintf("%s(%d)", dataType, charMaxLength.Int64)
		} else if numPrecision.Valid && numPrecision.Int64 > 0 {
			fullType = fmt.Sprintf("%s(%d)", dataType, numPrecision.Int64)
		}

		// Create the argument
		arg := schema.FunctionArg{
			Type: fullType,
		}

		// Set name if present
		if name.Valid && name.String != "" {
			arg.Name = name.String
		}

		// Add the argument to the function
		function.Arguments = append(function.Arguments, arg)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating function parameters: %w", err)
	}

	return nil
}
