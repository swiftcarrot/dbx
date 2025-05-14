package sqlite

import (
	"database/sql"
	"regexp"
	"strings"

	"github.com/swiftcarrot/dbx/schema"
)

// InspectTriggers retrieves all triggers from the database
func (s *SQLite) InspectTriggers(db *sql.DB, schm *schema.Schema) error {
	query := `
		SELECT name, tbl_name, sql FROM sqlite_master
		WHERE type = 'trigger'
		ORDER BY name
	`

	rows, err := db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var name, tableName, definition string
		if err := rows.Scan(&name, &tableName, &definition); err != nil {
			return err
		}

		// Parse trigger properties from the definition
		timing, events, forEach, when, _ := parseTriggerDefinition(definition)

		// Create the trigger object
		trigger := &schema.Trigger{
			Name:      name,
			Table:     tableName,
			Events:    events,
			Timing:    timing,
			ForEach:   forEach,
			When:      when,
			Function:  "", // SQLite doesn't use separate function objects, logic is in trigger body
			Arguments: []string{},
		}

		schm.Triggers = append(schm.Triggers, trigger)
	}

	return rows.Err()
}

// parseTriggerDefinition extracts trigger properties from its SQL definition
func parseTriggerDefinition(sql string) (timing string, events []string, forEach string, when string, body string) {
	sql = strings.TrimSpace(sql)

	// Extract timing (BEFORE, AFTER, INSTEAD OF)
	if strings.Contains(sql, "BEFORE ") {
		timing = "BEFORE"
	} else if strings.Contains(sql, "AFTER ") {
		timing = "AFTER"
	} else if strings.Contains(sql, "INSTEAD OF ") {
		timing = "INSTEAD OF"
	}

	// Extract events (INSERT, UPDATE, DELETE)
	events = []string{}
	if strings.Contains(sql, " INSERT ") || strings.Contains(sql, " INSERT\n") {
		events = append(events, "INSERT")
	}
	if strings.Contains(sql, " UPDATE ") || strings.Contains(sql, " UPDATE\n") ||
		strings.Contains(sql, " UPDATE OF ") {
		events = append(events, "UPDATE")
	}
	if strings.Contains(sql, " DELETE ") || strings.Contains(sql, " DELETE\n") {
		events = append(events, "DELETE")
	}

	// Extract FOR EACH ROW/STATEMENT
	if strings.Contains(sql, "FOR EACH ROW") {
		forEach = "ROW"
	} else {
		forEach = "STATEMENT"
	}

	// Extract WHEN condition if present
	whenPattern := regexp.MustCompile(`WHEN\s+\((.*?)\)`)
	whenMatches := whenPattern.FindStringSubmatch(sql)
	if len(whenMatches) > 1 {
		when = whenMatches[1]
	}

	// Extract trigger body
	bodyPattern := regexp.MustCompile(`BEGIN\s+(.*?)\s+END`)
	bodyMatches := bodyPattern.FindStringSubmatch(sql)
	if len(bodyMatches) > 1 {
		body = bodyMatches[1]
	}

	return
}
