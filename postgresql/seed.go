package postgresql

import (
	"database/sql"
	"fmt"
	"io"
	"strings"

	"github.com/swiftcarrot/dbx/seed"
)

// ImportFromCSV imports data from a CSV file into a PostgreSQL table
// It uses PostgreSQL's COPY command for efficient bulk data loading
func (pg *PostgreSQL) ImportFromCSV(db *sql.DB, tableName, schemaName string, reader io.Reader, options *seed.CSVImportOptions) error {
	// Set default options if not provided
	if options == nil {
		options = &CSVImportOptions{
			Delimiter: ",",
			NullValue: "",
			Header:    true,
			Quote:     "\"",
			Escape:    "\\",
			Encoding:  "UTF8",
		}
	}

	// Apply defaults for unset options
	if options.Delimiter == "" {
		options.Delimiter = ","
	}
	if options.Quote == "" {
		options.Quote = "\""
	}
	if options.Escape == "" {
		options.Escape = "\\"
	}
	if options.Encoding == "" {
		options.Encoding = "UTF8"
	}

	// Qualify the table name with schema if provided
	qualifiedTable := tableName
	if schemaName != "" && schemaName != "public" {
		qualifiedTable = fmt.Sprintf("%s.%s", quoteIdentifier(schemaName), quoteIdentifier(tableName))
	} else {
		qualifiedTable = quoteIdentifier(tableName)
	}

	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Build the COPY command
	copyCmd := fmt.Sprintf("COPY %s", qualifiedTable)

	// Add columns if specified
	if len(options.Columns) > 0 {
		columnNames := make([]string, len(options.Columns))
		for i, col := range options.Columns {
			columnNames[i] = quoteIdentifier(col)
		}
		copyCmd += fmt.Sprintf(" (%s)", strings.Join(columnNames, ", "))
	}

	copyCmd += " FROM STDIN WITH ("
	copyCmd += fmt.Sprintf("FORMAT CSV, DELIMITER '%s'", options.Delimiter)

	if options.Header {
		copyCmd += ", HEADER"
	}

	if options.Quote != "" {
		copyCmd += fmt.Sprintf(", QUOTE '%s'", options.Quote)
	}

	if options.Escape != "" {
		copyCmd += fmt.Sprintf(", ESCAPE '%s'", options.Escape)
	}

	if options.NullValue != "" {
		copyCmd += fmt.Sprintf(", NULL '%s'", options.NullValue)
	}

	if options.Encoding != "" {
		copyCmd += fmt.Sprintf(", ENCODING '%s'", options.Encoding)
	}

	copyCmd += ")"

	// Get the PostgreSQL db/sql connection
	stmt, err := tx.Prepare(copyCmd)
	if err != nil {
		return fmt.Errorf("failed to prepare COPY statement: %w", err)
	}
	defer stmt.Close()

	// Get a reference to the underlying PostgreSQL connection
	_, err = stmt.Exec()
	if err != nil {
		return fmt.Errorf("failed to execute COPY command: %w", err)
	}

	// Copy data
	_, err = io.Copy(tx.(*sql.Tx), reader)
	if err != nil {
		return fmt.Errorf("failed to copy data: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
