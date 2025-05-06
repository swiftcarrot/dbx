package postgresql

import (
	"database/sql"
	"strings"

	"github.com/swiftcarrot/dbx/schema"
)

// InspectTriggers retrieves all triggers from the database
func (pg *PostgreSQL) InspectTriggers(db *sql.DB, s *schema.Schema) error {
	query := `
		SELECT
			n.nspname AS schema_name,
			c.relname AS table_name,
			t.tgname AS trigger_name,
			pg_get_triggerdef(t.oid) AS definition
		FROM
			pg_trigger t
		JOIN
			pg_class c ON t.tgrelid = c.oid
		JOIN
			pg_namespace n ON c.relnamespace = n.oid
		WHERE
			NOT t.tgisinternal AND
			n.nspname NOT IN ('pg_catalog', 'information_schema')
		ORDER BY
			n.nspname, c.relname, t.tgname
	`

	rows, err := db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var schemaName, tableName, triggerName, definition string

		if err := rows.Scan(&schemaName, &tableName, &triggerName, &definition); err != nil {
			return err
		}

		// Parse the trigger definition to extract components
		// Example: CREATE TRIGGER trigger_name BEFORE INSERT ON table_name FOR EACH ROW EXECUTE FUNCTION function_name()

		timing := "BEFORE" // Default
		if strings.Contains(definition, "AFTER") {
			timing = "AFTER"
		} else if strings.Contains(definition, "INSTEAD OF") {
			timing = "INSTEAD OF"
		}

		forEach := "ROW" // Default
		if strings.Contains(definition, "FOR EACH STATEMENT") {
			forEach = "STATEMENT"
		}

		// Extract events (INSERT, UPDATE, DELETE)
		var events []string
		if strings.Contains(definition, " INSERT ") {
			events = append(events, "INSERT")
		}
		if strings.Contains(definition, " UPDATE ") {
			events = append(events, "UPDATE")
		}
		if strings.Contains(definition, " DELETE ") {
			events = append(events, "DELETE")
		}

		// Extract WHEN condition - properly handling nested parentheses
		var when string
		whenParts := strings.Split(definition, "WHEN (")
		if len(whenParts) > 1 {
			// Count parentheses to find the matching closing one
			whenClause := whenParts[1]
			openParens := 1 // We start with one open parenthesis from the "WHEN ("
			closeIndex := -1

			for i, char := range whenClause {
				if char == '(' {
					openParens++
				} else if char == ')' {
					openParens--
					if openParens == 0 {
						closeIndex = i
						break
					}
				}
			}

			if closeIndex >= 0 {
				when = whenClause[:closeIndex]
			}
		}

		// Extract function name and arguments
		functionInfo := ""
		execParts := strings.Split(definition, "EXECUTE ")
		if len(execParts) > 1 {
			// Check if PROCEDURE or FUNCTION is used (depends on PostgreSQL version)
			if strings.Contains(execParts[1], "PROCEDURE ") {
				functionInfo = strings.Split(execParts[1], "PROCEDURE ")[1]
			} else if strings.Contains(execParts[1], "FUNCTION ") {
				functionInfo = strings.Split(execParts[1], "FUNCTION ")[1]
			}
		}

		// Extract function name and arguments from functionInfo
		functionName := functionInfo
		var functionArgs []string

		// Remove trailing semicolon if present
		functionName = strings.TrimSuffix(functionName, ";")

		// Split function name and arguments
		if strings.Contains(functionName, "(") {
			parts := strings.SplitN(functionName, "(", 2)
			functionName = strings.TrimSpace(parts[0])

			if len(parts) > 1 {
				argsPart := strings.TrimSuffix(parts[1], ")")
				// Parse arguments, handling quoted strings properly
				if argsPart != "" {
					// This is a simplified parser that may need enhancement for complex cases
					var currentArg strings.Builder
					inQuotes := false

					for _, c := range argsPart {
						if c == '\'' {
							inQuotes = !inQuotes
							currentArg.WriteRune(c)
						} else if c == ',' && !inQuotes {
							functionArgs = append(functionArgs, strings.TrimSpace(currentArg.String()))
							currentArg.Reset()
						} else {
							currentArg.WriteRune(c)
						}
					}

					if currentArg.Len() > 0 {
						functionArgs = append(functionArgs, strings.TrimSpace(currentArg.String()))
					}
				}
			}
		}

		// Create the trigger
		trigger := s.CreateTrigger(
			triggerName,
			tableName,
			functionName,
			schema.TriggerInSchema(schemaName),
			schema.OnEvents(events...),
		)

		// Set timing
		switch timing {
		case "BEFORE":
			schema.Before(trigger)
		case "AFTER":
			schema.After(trigger)
		case "INSTEAD OF":
			schema.InsteadOf(trigger)
		}

		// Set scope
		if forEach == "ROW" {
			schema.ForEachRow(trigger)
		} else {
			schema.ForEachStatement(trigger)
		}

		// Set condition if present
		if when != "" {
			schema.WithCondition(when)(trigger)
		}

		// Set arguments if present
		if len(functionArgs) > 0 {
			schema.WithArguments(functionArgs...)(trigger)
		}
	}

	return rows.Err()
}
