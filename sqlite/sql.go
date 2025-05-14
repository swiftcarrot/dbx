package sqlite

import (
	"fmt"
	"strings"

	"github.com/swiftcarrot/dbx/schema"
)

// GenerateSQL generates SQL statements for SQLite database changes
func (s *SQLite) GenerateSQL(change schema.Change) (string, error) {
	switch c := change.(type) {
	case schema.CreateTableChange:
		return s.createTable(c)
	case schema.DropTableChange:
		return s.dropTable(c)
	case schema.AddColumnChange:
		return s.addColumn(c)
	case schema.DropColumnChange:
		return s.dropColumn(c)
	case schema.AlterColumnChange:
		return s.alterColumn(c)
	case schema.AddPrimaryKeyChange:
		return s.addPrimaryKey(c)
	case schema.DropPrimaryKeyChange:
		return s.dropPrimaryKey(c)
	case schema.AddIndexChange:
		return s.addIndex(c)
	case schema.DropIndexChange:
		return s.dropIndex(c)
	case schema.AddForeignKeyChange:
		return s.addForeignKey(c)
	case schema.DropForeignKeyChange:
		return s.dropForeignKey(c)
	case schema.CreateViewChange:
		return s.createView(c)
	case schema.AlterViewChange:
		return s.alterView(c)
	case schema.DropViewChange:
		return s.dropView(c)
	case schema.CreateTriggerChange:
		return s.createTrigger(c)
	case schema.AlterTriggerChange:
		return s.alterTrigger(c)
	case schema.DropTriggerChange:
		return s.dropTrigger(c)
	default:
		return "", fmt.Errorf("unsupported change type: %T", change)
	}
}

// createTable generates SQL for creating a table
func (s *SQLite) createTable(change schema.CreateTableChange) (string, error) {
	table := change.TableDef
	if table == nil {
		return "", fmt.Errorf("table definition is nil")
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("CREATE TABLE %s (\n", quoteIdentifier(table.Name)))

	// Add columns
	for i, col := range table.Columns {
		b.WriteString(fmt.Sprintf("  %s %s", quoteIdentifier(col.Name), col.Type))

		if !col.Nullable {
			b.WriteString(" NOT NULL")
		}

		if col.Default != "" {
			b.WriteString(fmt.Sprintf(" DEFAULT %s", col.Default))
		}

		if i < len(table.Columns)-1 || table.PrimaryKey != nil || len(table.ForeignKeys) > 0 {
			b.WriteString(",\n")
		}
	}

	// Add primary key
	if table.PrimaryKey != nil && len(table.PrimaryKey.Columns) > 0 {
		b.WriteString(fmt.Sprintf("  CONSTRAINT %s PRIMARY KEY (", quoteIdentifier(table.PrimaryKey.Name)))
		for i, col := range table.PrimaryKey.Columns {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(quoteIdentifier(col))
		}
		b.WriteString(")")

		if len(table.ForeignKeys) > 0 {
			b.WriteString(",\n")
		}
	}

	// Add foreign keys
	for i, fk := range table.ForeignKeys {
		b.WriteString(fmt.Sprintf("  CONSTRAINT %s FOREIGN KEY (", quoteIdentifier(fk.Name)))
		for j, col := range fk.Columns {
			if j > 0 {
				b.WriteString(", ")
			}
			b.WriteString(quoteIdentifier(col))
		}
		b.WriteString(fmt.Sprintf(") REFERENCES %s (", quoteIdentifier(fk.RefTable)))
		for j, col := range fk.RefColumns {
			if j > 0 {
				b.WriteString(", ")
			}
			b.WriteString(quoteIdentifier(col))
		}
		b.WriteString(")")

		if fk.OnDelete != "" {
			b.WriteString(fmt.Sprintf(" ON DELETE %s", fk.OnDelete))
		}
		if fk.OnUpdate != "" {
			b.WriteString(fmt.Sprintf(" ON UPDATE %s", fk.OnUpdate))
		}

		if i < len(table.ForeignKeys)-1 {
			b.WriteString(",\n")
		}
	}

	b.WriteString("\n);")

	return b.String(), nil
}

// dropTable generates SQL for dropping a table
func (s *SQLite) dropTable(change schema.DropTableChange) (string, error) {
	return fmt.Sprintf("DROP TABLE %s;", quoteIdentifier(change.TableName)), nil
}

// addColumn generates SQL for adding a column to a table
func (s *SQLite) addColumn(change schema.AddColumnChange) (string, error) {
	col := change.Column
	if col == nil {
		return "", fmt.Errorf("column definition is nil")
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s",
		quoteIdentifier(change.TableName),
		quoteIdentifier(col.Name),
		col.Type))

	if !col.Nullable {
		b.WriteString(" NOT NULL")
	}

	if col.Default != "" {
		b.WriteString(fmt.Sprintf(" DEFAULT %s", col.Default))
	}

	b.WriteString(";")
	return b.String(), nil
}

// dropColumn generates SQL for dropping a column from a table
// Note: SQLite does not support DROP COLUMN directly, it requires recreating the table
func (s *SQLite) dropColumn(change schema.DropColumnChange) (string, error) {
	return "", fmt.Errorf("SQLite does not support DROP COLUMN directly; you need to recreate the table")
}

// alterColumn generates SQL for altering a column
// Note: SQLite does not support ALTER COLUMN directly, it requires recreating the table
func (s *SQLite) alterColumn(change schema.AlterColumnChange) (string, error) {
	return "", fmt.Errorf("SQLite does not support ALTER COLUMN directly; you need to recreate the table")
}

// addPrimaryKey generates SQL for adding a primary key
// Note: SQLite does not support adding primary keys to existing tables
func (s *SQLite) addPrimaryKey(change schema.AddPrimaryKeyChange) (string, error) {
	return "", fmt.Errorf("SQLite does not support adding primary keys to existing tables; you need to recreate the table")
}

// dropPrimaryKey generates SQL for dropping a primary key
// Note: SQLite does not support dropping primary keys
func (s *SQLite) dropPrimaryKey(change schema.DropPrimaryKeyChange) (string, error) {
	return "", fmt.Errorf("SQLite does not support dropping primary keys; you need to recreate the table")
}

// addIndex generates SQL for adding an index
func (s *SQLite) addIndex(change schema.AddIndexChange) (string, error) {
	idx := change.Index
	if idx == nil {
		return "", fmt.Errorf("index definition is nil")
	}

	var b strings.Builder
	if idx.Unique {
		b.WriteString("CREATE UNIQUE INDEX ")
	} else {
		b.WriteString("CREATE INDEX ")
	}

	b.WriteString(fmt.Sprintf("%s ON %s (",
		quoteIdentifier(idx.Name),
		quoteIdentifier(change.TableName)))

	for i, col := range idx.Columns {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(quoteIdentifier(col))
	}
	b.WriteString(");")

	return b.String(), nil
}

// dropIndex generates SQL for dropping an index
func (s *SQLite) dropIndex(change schema.DropIndexChange) (string, error) {
	return fmt.Sprintf("DROP INDEX %s;", quoteIdentifier(change.IndexName)), nil
}

// addForeignKey generates SQL for adding a foreign key
// Note: SQLite only supports foreign keys when creating tables
func (s *SQLite) addForeignKey(change schema.AddForeignKeyChange) (string, error) {
	return "", fmt.Errorf("SQLite does not support adding foreign keys to existing tables; you need to recreate the table")
}

// dropForeignKey generates SQL for dropping a foreign key
// Note: SQLite does not support dropping foreign keys
func (s *SQLite) dropForeignKey(change schema.DropForeignKeyChange) (string, error) {
	return "", fmt.Errorf("SQLite does not support dropping foreign keys; you need to recreate the table")
}

// createView generates SQL for creating a view
func (s *SQLite) createView(change schema.CreateViewChange) (string, error) {
	view := change.View
	if view == nil {
		return "", fmt.Errorf("view definition is nil")
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("CREATE VIEW %s", quoteIdentifier(view.Name)))

	if len(view.Columns) > 0 {
		b.WriteString(" (")
		for i, col := range view.Columns {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(quoteIdentifier(col))
		}
		b.WriteString(")")
	}

	b.WriteString(" AS ")
	b.WriteString(view.Definition)
	b.WriteString(";")

	return b.String(), nil
}

// alterView generates SQL for altering a view
func (s *SQLite) alterView(change schema.AlterViewChange) (string, error) {
	view := change.View
	if view == nil {
		return "", fmt.Errorf("view definition is nil")
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("DROP VIEW IF EXISTS %s;\n", quoteIdentifier(view.Name)))
	b.WriteString(fmt.Sprintf("CREATE VIEW %s", quoteIdentifier(view.Name)))

	if len(view.Columns) > 0 {
		b.WriteString(" (")
		for i, col := range view.Columns {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(quoteIdentifier(col))
		}
		b.WriteString(")")
	}

	b.WriteString(" AS ")
	b.WriteString(view.Definition)
	b.WriteString(";")

	return b.String(), nil
}

// dropView generates SQL for dropping a view
func (s *SQLite) dropView(change schema.DropViewChange) (string, error) {
	return fmt.Sprintf("DROP VIEW %s;", quoteIdentifier(change.ViewName)), nil
}

// createTrigger generates SQL for creating a trigger
func (s *SQLite) createTrigger(change schema.CreateTriggerChange) (string, error) {
	trigger := change.Trigger
	if trigger == nil {
		return "", fmt.Errorf("trigger definition is nil")
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("CREATE TRIGGER %s\n", quoteIdentifier(trigger.Name)))

	b.WriteString(trigger.Timing)
	b.WriteString(" ")

	for i, event := range trigger.Events {
		if i > 0 {
			b.WriteString(" OR ")
		}
		b.WriteString(event)
	}

	b.WriteString(fmt.Sprintf(" ON %s\n", quoteIdentifier(trigger.Table)))
	b.WriteString(fmt.Sprintf("FOR EACH %s\n", trigger.ForEach))

	if trigger.When != "" {
		b.WriteString(fmt.Sprintf("WHEN (%s)\n", trigger.When))
	}

	b.WriteString("BEGIN\n")

	// In SQLite, we need to include the trigger logic in the trigger body instead of calling a function
	if strings.TrimSpace(trigger.Function) != "" {
		b.WriteString(fmt.Sprintf("  SELECT %s", trigger.Function))

		if len(trigger.Arguments) > 0 {
			b.WriteString("(")
			for i, arg := range trigger.Arguments {
				if i > 0 {
					b.WriteString(", ")
				}
				b.WriteString(arg)
			}
			b.WriteString(")")
		}

		b.WriteString(";\n")
	}

	b.WriteString("END;")

	return b.String(), nil
}

// alterTrigger generates SQL for altering a trigger
func (s *SQLite) alterTrigger(change schema.AlterTriggerChange) (string, error) {
	trigger := change.Trigger
	if trigger == nil {
		return "", fmt.Errorf("trigger definition is nil")
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("DROP TRIGGER IF EXISTS %s;\n", quoteIdentifier(trigger.Name)))

	b.WriteString(fmt.Sprintf("CREATE TRIGGER %s\n", quoteIdentifier(trigger.Name)))

	b.WriteString(trigger.Timing)
	b.WriteString(" ")

	for i, event := range trigger.Events {
		if i > 0 {
			b.WriteString(" OR ")
		}
		b.WriteString(event)
	}

	b.WriteString(fmt.Sprintf(" ON %s\n", quoteIdentifier(trigger.Table)))
	b.WriteString(fmt.Sprintf("FOR EACH %s\n", trigger.ForEach))

	if trigger.When != "" {
		b.WriteString(fmt.Sprintf("WHEN (%s)\n", trigger.When))
	}

	b.WriteString("BEGIN\n")

	// In SQLite, we include the trigger logic in the trigger body
	if strings.TrimSpace(trigger.Function) != "" {
		b.WriteString(fmt.Sprintf("  SELECT %s", trigger.Function))

		if len(trigger.Arguments) > 0 {
			b.WriteString("(")
			for i, arg := range trigger.Arguments {
				if i > 0 {
					b.WriteString(", ")
				}
				b.WriteString(arg)
			}
			b.WriteString(")")
		}

		b.WriteString(";\n")
	}

	b.WriteString("END;")

	return b.String(), nil
}

// dropTrigger generates SQL for dropping a trigger
func (s *SQLite) dropTrigger(change schema.DropTriggerChange) (string, error) {
	return fmt.Sprintf("DROP TRIGGER %s;", quoteIdentifier(change.TriggerName)), nil
}

// Helper functions for quoting identifiers and literals
func quoteIdentifier(s string) string {
	// Check if it's already quoted
	if strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"") {
		return s
	}

	// Handle schema-qualified identifiers
	parts := strings.Split(s, ".")
	for i, part := range parts {
		if !strings.HasPrefix(part, "\"") && !strings.HasSuffix(part, "\"") {
			parts[i] = "\"" + part + "\""
		}
	}
	return strings.Join(parts, ".")
}

func quoteLiteral(s string) string {
	// Check if it's already quoted
	if strings.HasPrefix(s, "'") && strings.HasSuffix(s, "'") {
		return s
	}

	// Escape single quotes
	s = strings.ReplaceAll(s, "'", "''")
	return "'" + s + "'"
}
