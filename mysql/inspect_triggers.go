package mysql

import (
	"database/sql"
	"fmt"

	"github.com/swiftcarrot/dbx/schema"
)

// InspectTriggers inspects triggers in the database
func (my *MySQL) InspectTriggers(db *sql.DB, s *schema.Schema) error {
	// Query to get triggers
	query := `
		SELECT
			trigger_name,
			event_manipulation,
			action_timing,
			event_object_table,
			action_statement
		FROM
			information_schema.triggers
		WHERE
			trigger_schema = DATABASE()
		ORDER BY
			trigger_name;
	`

	rows, err := db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			name       string
			event      string
			timing     string
			table      string
			actionStmt string
		)

		if err := rows.Scan(&name, &event, &timing, &table, &actionStmt); err != nil {
			return err
		}

		// Create the trigger
		trigger := &schema.Trigger{
			Name:     name,
			Table:    table,
			Events:   []string{event}, // MySQL triggers have one event type per trigger
			Timing:   timing,
			ForEach:  "ROW", // MySQL triggers are always FOR EACH ROW
			Function: "",    // MySQL triggers don't reference functions, just have direct action statements
		}

		// In MySQL, the trigger action is directly in the trigger definition
		// There's no separate function to call, so we'll store it in the Function field
		// for now, even though it's not a function name but the actual SQL statements
		trigger.Function = actionStmt

		// Add the trigger to the schema
		s.Triggers = append(s.Triggers, trigger)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating triggers: %w", err)
	}

	return nil
}
