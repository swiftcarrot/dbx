package postgresql

import (
	"fmt"
	"strings"

	"github.com/swiftcarrot/dbx/schema"
)

// GenerateSQL converts a schema change to a PostgreSQL SQL statement
func (pg *PostgreSQL) GenerateSQL(change schema.Change) (string, error) {
	switch c := change.(type) {
	// Schema-related changes
	case schema.CreateSchemaChange:
		return pg.generateCreateSchema(c), nil
	case schema.DropSchemaChange:
		return pg.generateDropSchema(c), nil

	// Extension-related changes
	case schema.EnableExtensionChange:
		return pg.generateEnableExtension(c), nil
	case schema.DisableExtensionChange:
		return pg.generateDisableExtension(c), nil

	// Table-related changes
	case schema.CreateTableChange:
		return pg.generateCreateTable(c), nil
	case schema.DropTableChange:
		return pg.generateDropTable(c), nil

	// Column-related changes
	case schema.AddColumnChange:
		return pg.generateAddColumn(c), nil
	case schema.DropColumnChange:
		return pg.generateDropColumn(c), nil
	case schema.AlterColumnChange:
		return pg.generateAlterColumn(c), nil

	// Primary key-related changes
	case schema.AddPrimaryKeyChange:
		return pg.generateAddPrimaryKey(c), nil
	case schema.DropPrimaryKeyChange:
		return pg.generateDropPrimaryKey(c), nil

	// Index-related changes
	case schema.AddIndexChange:
		return pg.generateAddIndex(c), nil
	case schema.DropIndexChange:
		return pg.generateDropIndex(c), nil

	// Foreign key-related changes
	case schema.AddForeignKeyChange:
		return pg.generateAddForeignKey(c), nil
	case schema.DropForeignKeyChange:
		return pg.generateDropForeignKey(c), nil

	// Sequence-related changes
	case schema.CreateSequenceChange:
		return pg.generateCreateSequence(c), nil
	case schema.AlterSequenceChange:
		return pg.generateAlterSequence(c), nil
	case schema.DropSequenceChange:
		return pg.generateDropSequence(c), nil

	// Function-related changes
	case schema.CreateFunctionChange:
		return pg.generateCreateFunction(c), nil
	case schema.AlterFunctionChange:
		return pg.generateAlterFunction(c), nil
	case schema.DropFunctionChange:
		return pg.generateDropFunction(c), nil

	// View-related changes
	case schema.CreateViewChange:
		return pg.generateCreateView(c), nil
	case schema.AlterViewChange:
		return pg.generateAlterView(c), nil
	case schema.DropViewChange:
		return pg.generateDropView(c), nil

	// Trigger-related changes
	case schema.CreateTriggerChange:
		return pg.generateCreateTrigger(c), nil
	case schema.AlterTriggerChange:
		return pg.generateAlterTrigger(c), nil
	case schema.DropTriggerChange:
		return pg.generateDropTrigger(c), nil

	default:
		return "", fmt.Errorf("unsupported change type: %T", change)
	}
}

// Schema-related SQL generation

func (pg *PostgreSQL) generateCreateSchema(c schema.CreateSchemaChange) string {
	return fmt.Sprintf("CREATE SCHEMA %s;", quoteIdentifier(c.SchemaName))
}

func (pg *PostgreSQL) generateDropSchema(c schema.DropSchemaChange) string {
	return fmt.Sprintf("DROP SCHEMA %s;", quoteIdentifier(c.SchemaName))
}

// Extension-related SQL generation

func (pg *PostgreSQL) generateEnableExtension(c schema.EnableExtensionChange) string {
	return fmt.Sprintf("CREATE EXTENSION IF NOT EXISTS %s;", quoteIdentifier(c.Extension))
}

func (pg *PostgreSQL) generateDisableExtension(c schema.DisableExtensionChange) string {
	return fmt.Sprintf("DROP EXTENSION IF EXISTS %s;", quoteIdentifier(c.Extension))
}

// Table-related SQL generation

func (pg *PostgreSQL) generateCreateTable(c schema.CreateTableChange) string {
	table := c.TableDef
	var sb strings.Builder

	tableName := table.Name
	if table.Schema != "" && table.Schema != "public" {
		tableName = table.Schema + "." + tableName
	}

	sb.WriteString(fmt.Sprintf("CREATE TABLE %s (\n", quoteIdentifier(tableName)))

	// Add columns
	for i, col := range table.Columns {
		if i > 0 {
			sb.WriteString(",\n")
		}
		sb.WriteString(fmt.Sprintf("  %s %s", quoteIdentifier(col.Name), col.TypeSQL()))

		// Add NOT NULL constraint if needed
		if !col.Nullable {
			sb.WriteString(" NOT NULL")
		}

		// Add default value if specified
		if col.Default != "" {
			sb.WriteString(fmt.Sprintf(" DEFAULT %s", col.Default))
		}
	}

	// Add primary key constraint directly in the CREATE TABLE statement
	if table.PrimaryKey != nil {
		pk := table.PrimaryKey
		pkColumnsList := make([]string, len(pk.Columns))
		for i, col := range pk.Columns {
			pkColumnsList[i] = quoteIdentifier(col)
		}
		sb.WriteString(fmt.Sprintf(",\n  CONSTRAINT %s PRIMARY KEY (%s)",
			quoteIdentifier(pk.Name),
			strings.Join(pkColumnsList, ", ")))
	}

	sb.WriteString("\n);")

	// Add comments for columns if present
	for _, col := range table.Columns {
		if col.Comment != "" {
			sb.WriteString(fmt.Sprintf("\nCOMMENT ON COLUMN %s.%s IS %s;",
				quoteIdentifier(tableName),
				quoteIdentifier(col.Name),
				quoteLiteral(col.Comment)))
		}
	}

	return sb.String()
}

func (pg *PostgreSQL) generateDropTable(c schema.DropTableChange) string {
	tableName := c.TableName
	if c.SchemaName != "" && c.SchemaName != "public" {
		tableName = c.SchemaName + "." + tableName
	}
	return fmt.Sprintf("DROP TABLE %s;", quoteIdentifier(tableName))
}

// Column-related SQL generation

func (pg *PostgreSQL) generateAddColumn(c schema.AddColumnChange) string {
	column := c.Column
	sql := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s",
		quoteIdentifier(c.TableName),
		quoteIdentifier(column.Name),
		column.TypeSQL())

	if !column.Nullable {
		sql += " NOT NULL"
	}

	if column.Default != "" {
		sql += fmt.Sprintf(" DEFAULT %s", column.Default)
	}

	sql += ";"

	// Add comment if present
	if column.Comment != "" {
		sql += fmt.Sprintf("\nCOMMENT ON COLUMN %s.%s IS %s;",
			quoteIdentifier(c.TableName),
			quoteIdentifier(column.Name),
			quoteLiteral(column.Comment))
	}

	return sql
}

func (pg *PostgreSQL) generateDropColumn(c schema.DropColumnChange) string {
	return fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s;",
		quoteIdentifier(c.TableName),
		quoteIdentifier(c.ColumnName))
}

func (pg *PostgreSQL) generateAlterColumn(c schema.AlterColumnChange) string {
	column := c.Column
	var statements []string

	// Type change
	statements = append(statements, fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s TYPE %s;",
		quoteIdentifier(c.TableName),
		quoteIdentifier(column.Name),
		column.TypeSQL()))

	// Nullability change
	if !column.Nullable {
		statements = append(statements, fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s SET NOT NULL;",
			quoteIdentifier(c.TableName),
			quoteIdentifier(column.Name)))
	} else {
		statements = append(statements, fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s DROP NOT NULL;",
			quoteIdentifier(c.TableName),
			quoteIdentifier(column.Name)))
	}

	// Default value change
	if column.Default != "" {
		statements = append(statements, fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s SET DEFAULT %s;",
			quoteIdentifier(c.TableName),
			quoteIdentifier(column.Name),
			column.Default))
	} else {
		statements = append(statements, fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s DROP DEFAULT;",
			quoteIdentifier(c.TableName),
			quoteIdentifier(column.Name)))
	}

	// Comment change
	if column.Comment != "" {
		statements = append(statements, fmt.Sprintf("COMMENT ON COLUMN %s.%s IS %s;",
			quoteIdentifier(c.TableName),
			quoteIdentifier(column.Name),
			quoteLiteral(column.Comment)))
	}

	return strings.Join(statements, "\n")
}

// Primary key-related SQL generation

func (pg *PostgreSQL) generateAddPrimaryKey(c schema.AddPrimaryKeyChange) string {
	pk := c.PrimaryKey
	pkColumnsList := make([]string, len(pk.Columns))
	for i, col := range pk.Columns {
		pkColumnsList[i] = quoteIdentifier(col)
	}

	return fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s PRIMARY KEY (%s);",
		quoteIdentifier(c.TableName),
		quoteIdentifier(pk.Name),
		strings.Join(pkColumnsList, ", "))
}

func (pg *PostgreSQL) generateDropPrimaryKey(c schema.DropPrimaryKeyChange) string {
	return fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s;",
		quoteIdentifier(c.TableName),
		quoteIdentifier(c.PKName))
}

// Index-related SQL generation

func (pg *PostgreSQL) generateAddIndex(c schema.AddIndexChange) string {
	idx := c.Index
	unique := ""
	if idx.Unique {
		unique = "UNIQUE "
	}

	columns := make([]string, len(idx.Columns))
	for i, col := range idx.Columns {
		columns[i] = quoteIdentifier(col)
	}

	return fmt.Sprintf("CREATE %sINDEX %s ON %s (%s);",
		unique,
		quoteIdentifier(idx.Name),
		quoteIdentifier(c.TableName),
		strings.Join(columns, ", "))
}

func (pg *PostgreSQL) generateDropIndex(c schema.DropIndexChange) string {
	return fmt.Sprintf("DROP INDEX %s;", quoteIdentifier(c.IndexName))
}

// Foreign key-related SQL generation

func (pg *PostgreSQL) generateAddForeignKey(c schema.AddForeignKeyChange) string {
	fk := c.ForeignKey
	srcColumns := make([]string, len(fk.Columns))
	refColumns := make([]string, len(fk.RefColumns))

	for i, col := range fk.Columns {
		srcColumns[i] = quoteIdentifier(col)
	}

	for i, col := range fk.RefColumns {
		refColumns[i] = quoteIdentifier(col)
	}

	sql := fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s (%s)",
		quoteIdentifier(c.TableName),
		quoteIdentifier(fk.Name),
		strings.Join(srcColumns, ", "),
		quoteIdentifier(fk.RefTable),
		strings.Join(refColumns, ", "))

	if fk.OnDelete != "" {
		sql += fmt.Sprintf(" ON DELETE %s", fk.OnDelete)
	}

	if fk.OnUpdate != "" {
		sql += fmt.Sprintf(" ON UPDATE %s", fk.OnUpdate)
	}

	return sql + ";"
}

func (pg *PostgreSQL) generateDropForeignKey(c schema.DropForeignKeyChange) string {
	return fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s;",
		quoteIdentifier(c.TableName),
		quoteIdentifier(c.FKName))
}

// Sequence-related SQL generation

func (pg *PostgreSQL) generateCreateSequence(c schema.CreateSequenceChange) string {
	seq := c.Sequence
	var sb strings.Builder

	sequenceName := seq.Name
	if seq.Schema != "" && seq.Schema != "public" {
		sequenceName = seq.Schema + "." + sequenceName
	}

	sb.WriteString(fmt.Sprintf("CREATE SEQUENCE %s", quoteIdentifier(sequenceName)))

	if seq.Increment != 1 {
		sb.WriteString(fmt.Sprintf(" INCREMENT BY %d", seq.Increment))
	}

	if seq.MinValue != 1 {
		sb.WriteString(fmt.Sprintf(" MINVALUE %d", seq.MinValue))
	}

	if seq.MaxValue != 9223372036854775807 { // Default PostgreSQL bigint max
		sb.WriteString(fmt.Sprintf(" MAXVALUE %d", seq.MaxValue))
	}

	if seq.Start != 1 {
		sb.WriteString(fmt.Sprintf(" START WITH %d", seq.Start))
	}

	if seq.Cache != 1 {
		sb.WriteString(fmt.Sprintf(" CACHE %d", seq.Cache))
	}

	if seq.Cycle {
		sb.WriteString(" CYCLE")
	}

	sb.WriteString(";")
	return sb.String()
}

func (pg *PostgreSQL) generateAlterSequence(c schema.AlterSequenceChange) string {
	seq := c.Sequence
	var sb strings.Builder

	sequenceName := seq.Name
	if seq.Schema != "" && seq.Schema != "public" {
		sequenceName = seq.Schema + "." + sequenceName
	}

	sb.WriteString(fmt.Sprintf("ALTER SEQUENCE %s", quoteIdentifier(sequenceName)))

	// We only include properties that make sense to alter
	if seq.Increment != 1 {
		sb.WriteString(fmt.Sprintf(" INCREMENT BY %d", seq.Increment))
	}

	if seq.MinValue != 1 {
		sb.WriteString(fmt.Sprintf(" MINVALUE %d", seq.MinValue))
	} else {
		sb.WriteString(" NO MINVALUE")
	}

	if seq.MaxValue != 9223372036854775807 {
		sb.WriteString(fmt.Sprintf(" MAXVALUE %d", seq.MaxValue))
	} else {
		sb.WriteString(" NO MAXVALUE")
	}

	if seq.Cache != 1 {
		sb.WriteString(fmt.Sprintf(" CACHE %d", seq.Cache))
	}

	if seq.Cycle {
		sb.WriteString(" CYCLE")
	} else {
		sb.WriteString(" NO CYCLE")
	}

	sb.WriteString(";")
	return sb.String()
}

func (pg *PostgreSQL) generateDropSequence(c schema.DropSequenceChange) string {
	sequenceName := c.SequenceName
	if c.SchemaName != "" && c.SchemaName != "public" {
		sequenceName = c.SchemaName + "." + sequenceName
	}
	return fmt.Sprintf("DROP SEQUENCE %s;", quoteIdentifier(sequenceName))
}

// Function-related SQL generation

func (pg *PostgreSQL) generateCreateFunction(c schema.CreateFunctionChange) string {
	return pg.generateFunctionSQL("CREATE FUNCTION", c.Function)
}

func (pg *PostgreSQL) generateAlterFunction(c schema.AlterFunctionChange) string {
	return pg.generateFunctionSQL("CREATE OR REPLACE FUNCTION", c.Function)
}

func (pg *PostgreSQL) generateFunctionSQL(command string, fn *schema.Function) string {
	var sb strings.Builder

	functionName := fn.Name
	if fn.Schema != "" && fn.Schema != "public" {
		functionName = fn.Schema + "." + functionName
	}

	// Function declaration
	sb.WriteString(fmt.Sprintf("%s %s(", command, quoteIdentifier(functionName)))

	// Function arguments
	var args []string
	for _, arg := range fn.Arguments {
		argStr := ""
		if arg.Name != "" {
			argStr += arg.Name + " "
		}

		if arg.Mode != "IN" { // IN is the default so we don't need to specify it
			argStr += arg.Mode + " "
		}

		argStr += arg.Type

		if arg.Default != "" {
			argStr += fmt.Sprintf(" DEFAULT %s", arg.Default)
		}

		args = append(args, argStr)
	}
	sb.WriteString(strings.Join(args, ", "))
	sb.WriteString(")")

	// Return type
	sb.WriteString(fmt.Sprintf(" RETURNS %s AS $$", fn.Returns))

	// Function body
	sb.WriteString(fn.Body)
	sb.WriteString("$$ LANGUAGE ")
	sb.WriteString(fn.Language)

	// Function properties
	if fn.Volatility != "VOLATILE" {
		sb.WriteString(" " + fn.Volatility)
	}

	if fn.Strict {
		sb.WriteString(" STRICT")
	}

	if fn.Security != "INVOKER" {
		sb.WriteString(" SECURITY " + fn.Security)
	}

	sb.WriteString(fmt.Sprintf(" COST %d", fn.Cost))
	sb.WriteString(";")

	return sb.String()
}

func (pg *PostgreSQL) generateDropFunction(c schema.DropFunctionChange) string {
	functionName := c.FunctionName
	if c.SchemaName != "" && c.SchemaName != "public" {
		functionName = c.SchemaName + "." + functionName
	}

	// Build argument type list for overloaded functions
	var argTypes []string
	for _, arg := range c.FunctionArgs {
		argTypes = append(argTypes, arg.Type)
	}

	if len(argTypes) > 0 {
		return fmt.Sprintf("DROP FUNCTION %s(%s);",
			quoteIdentifier(functionName),
			strings.Join(argTypes, ", "))
	}

	return fmt.Sprintf("DROP FUNCTION %s;", quoteIdentifier(functionName))
}

// View-related SQL generation

func (pg *PostgreSQL) generateCreateView(c schema.CreateViewChange) string {
	view := c.View
	var sb strings.Builder

	viewName := view.Name
	if view.Schema != "" && view.Schema != "public" {
		viewName = view.Schema + "." + viewName
	}

	sb.WriteString(fmt.Sprintf("CREATE VIEW %s", quoteIdentifier(viewName)))

	// Optional column names
	if len(view.Columns) > 0 {
		columns := make([]string, len(view.Columns))
		for i, col := range view.Columns {
			columns[i] = quoteIdentifier(col)
		}
		sb.WriteString(fmt.Sprintf(" (%s)", strings.Join(columns, ", ")))
	}

	// Optional view options
	if len(view.Options) > 0 {
		sb.WriteString(fmt.Sprintf(" WITH (%s)", strings.Join(view.Options, ", ")))
	}

	// View definition
	sb.WriteString(" AS ")
	sb.WriteString(view.Definition)

	if !strings.HasSuffix(view.Definition, ";") {
		sb.WriteString(";")
	}

	return sb.String()
}

func (pg *PostgreSQL) generateAlterView(c schema.AlterViewChange) string {
	// For PostgreSQL, we create or replace the view rather than altering it
	return pg.generateCreateView(schema.CreateViewChange{View: c.View})
}

func (pg *PostgreSQL) generateDropView(c schema.DropViewChange) string {
	viewName := c.ViewName
	if c.SchemaName != "" && c.SchemaName != "public" {
		viewName = c.SchemaName + "." + viewName
	}
	return fmt.Sprintf("DROP VIEW %s;", quoteIdentifier(viewName))
}

// Trigger-related SQL generation

func (pg *PostgreSQL) generateCreateTrigger(c schema.CreateTriggerChange) string {
	return pg.generateTriggerSQL(c.Trigger)
}

func (pg *PostgreSQL) generateAlterTrigger(c schema.AlterTriggerChange) string {
	// PostgreSQL doesn't have a direct ALTER TRIGGER syntax for changing trigger definitions
	// So we drop and recreate it
	trigger := c.Trigger
	dropSQL := fmt.Sprintf("DROP TRIGGER %s ON %s;",
		quoteIdentifier(trigger.Name),
		quoteIdentifier(trigger.Table))
	createSQL := pg.generateTriggerSQL(trigger)

	return dropSQL + "\n" + createSQL
}

func (pg *PostgreSQL) generateTriggerSQL(trigger *schema.Trigger) string {
	var sb strings.Builder

	triggerName := trigger.Name
	if trigger.Schema != "" && trigger.Schema != "public" {
		triggerName = trigger.Schema + "." + triggerName
	}

	sb.WriteString(fmt.Sprintf("CREATE TRIGGER %s\n", quoteIdentifier(triggerName)))
	sb.WriteString(trigger.Timing + " ")
	sb.WriteString(strings.Join(trigger.Events, " OR "))
	sb.WriteString(fmt.Sprintf(" ON %s\n", quoteIdentifier(trigger.Table)))
	sb.WriteString(fmt.Sprintf("FOR EACH %s\n", trigger.ForEach))

	if trigger.When != "" {
		sb.WriteString("WHEN (" + trigger.When + ")\n")
	}

	sb.WriteString("EXECUTE FUNCTION " + trigger.Function)

	if len(trigger.Arguments) > 0 {
		sb.WriteString("(" + strings.Join(trigger.Arguments, ", ") + ")")
	} else {
		sb.WriteString("()")
	}

	sb.WriteString(";")
	return sb.String()
}

func (pg *PostgreSQL) generateDropTrigger(c schema.DropTriggerChange) string {
	triggerTable := c.TriggerTable
	if c.SchemaName != "" && c.SchemaName != "public" {
		triggerTable = c.SchemaName + "." + triggerTable
	}

	return fmt.Sprintf("DROP TRIGGER %s ON %s;",
		quoteIdentifier(c.TriggerName),
		quoteIdentifier(triggerTable))
}

// Helper functions for quoting identifiers and literals

func quoteIdentifier(s string) string {
	// Handle schema-qualified identifiers
	if strings.Contains(s, ".") {
		parts := strings.Split(s, ".")
		for i, part := range parts {
			parts[i] = quoteIdentifier(part)
		}
		return strings.Join(parts, ".")
	}

	// Don't double quote
	if strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"") {
		return s
	}

	return "\"" + strings.ReplaceAll(s, "\"", "\"\"") + "\""
}

func quoteLiteral(s string) string {
	// Don't double quote
	if strings.HasPrefix(s, "'") && strings.HasSuffix(s, "'") {
		return s
	}

	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

func PostgresArrayToSlice(pgArray string) []string {
	if len(pgArray) < 2 || pgArray[0] != '{' || pgArray[len(pgArray)-1] != '}' {
		return nil
	}

	trimmed := pgArray[1 : len(pgArray)-1]
	if trimmed == "" {
		return []string{}
	}

	result := strings.Split(trimmed, ",")
	for i, val := range result {
		result[i] = strings.TrimSpace(val)
	}

	return result
}
