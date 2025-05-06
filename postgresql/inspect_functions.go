package postgresql

import (
	"database/sql"
	"strings"

	"github.com/swiftcarrot/dbx/schema"
)

// InspectFunctions retrieves all functions from the database
func (pg *PostgreSQL) InspectFunctions(db *sql.DB, s *schema.Schema) error {
	query := `
		SELECT
			n.nspname AS schema_name,
			p.proname AS function_name,
			pg_get_function_result(p.oid) AS result_type,
			pg_get_function_arguments(p.oid) AS argument_types,
			pg_get_functiondef(p.oid) AS definition,
			l.lanname AS language,
			CASE
				WHEN p.provolatile = 'i' THEN 'IMMUTABLE'
				WHEN p.provolatile = 's' THEN 'STABLE'
				WHEN p.provolatile = 'v' THEN 'VOLATILE'
			END AS volatility,
			p.proisstrict AS strict,
			CASE
				WHEN p.prosecdef THEN 'DEFINER'
				ELSE 'INVOKER'
			END AS security,
			p.procost AS cost
		FROM
			pg_proc p
		JOIN
			pg_namespace n ON p.pronamespace = n.oid
		JOIN
			pg_language l ON p.prolang = l.oid
		WHERE
			n.nspname NOT IN ('pg_catalog', 'information_schema')
			AND p.prokind = 'f' -- Function (not procedure or aggregate)
		ORDER BY
			n.nspname, p.proname
	`

	rows, err := db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var schemaName, functionName, resultType, argumentStr, definition, language, volatility, security string
		var strict bool
		var cost int

		if err := rows.Scan(&schemaName, &functionName, &resultType, &argumentStr, &definition, &language, &volatility, &strict, &security, &cost); err != nil {
			return err
		}

		// Extract function body from definition
		// The definition includes CREATE OR REPLACE FUNCTION... and the body
		// We need to extract just the body part between $function$ and $function$
		body := ""
		parts := strings.Split(definition, "AS ")
		if len(parts) > 1 {
			// Find the content between the first pair of $function$ delimiters
			dollarParts := strings.Split(parts[1], "$function$")
			if len(dollarParts) > 2 {
				body = dollarParts[1]
			}
		}

		// Parse arguments string into FunctionArg structs
		var args []schema.FunctionArg
		if argumentStr != "" {
			// Example argumentStr: "arg1 integer, arg2 text DEFAULT 'test'"
			argParts := strings.Split(argumentStr, ",")
			for _, part := range argParts {
				part = strings.TrimSpace(part)
				if part == "" {
					continue
				}

				// Parse each argument with possible modes IN, OUT, INOUT, VARIADIC
				argFields := strings.Fields(part)
				if len(argFields) >= 2 {
					arg := schema.FunctionArg{
						Mode: "IN", // Default mode if not specified
					}

					// Check if the first token is a mode
					startIdx := 0
					if argFields[0] == "IN" || argFields[0] == "OUT" || argFields[0] == "INOUT" || argFields[0] == "VARIADIC" {
						arg.Mode = argFields[0]
						startIdx = 1
					}

					// The next token is the argument name (without the $ prefix that PostgreSQL adds for unnamed arguments)
					if len(argFields) > startIdx && !strings.HasPrefix(argFields[startIdx], "$") {
						arg.Name = argFields[startIdx]
						startIdx++
					}

					// Extract the type
					if len(argFields) > startIdx {
						arg.Type = argFields[startIdx]
						startIdx++
					}

					// Extract default value if present
					if strings.Contains(part, "DEFAULT") {
						defaultParts := strings.Split(part, "DEFAULT")
						if len(defaultParts) > 1 {
							arg.Default = strings.TrimSpace(defaultParts[1])
						}
					}

					args = append(args, arg)
				}
			}
		}

		function := s.CreateFunction(
			functionName,
			resultType,
			body,
			schema.Language(language),
			schema.FunctionCost(cost),
			schema.FunctionInSchema(schemaName),
			schema.FunctionArgs(args...),
		)

		// Set volatility
		switch volatility {
		case "IMMUTABLE":
			schema.Immutable(function)
		case "STABLE":
			schema.Stable(function)
		case "VOLATILE":
			schema.Volatile(function)
		}

		// Set strictness
		if strict {
			schema.Strict(function)
		} else {
			schema.NotStrict(function)
		}

		// Set security
		if security == "DEFINER" {
			schema.SecurityDefiner(function)
		} else {
			schema.SecurityInvoker(function)
		}
	}

	return rows.Err()
}
