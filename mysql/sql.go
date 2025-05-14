package mysql

import (
	"fmt"
	"strings"

	"github.com/swiftcarrot/dbx/schema"
)

// GenerateSQL converts a schema change to a MySQL SQL statement
func (my *MySQL) GenerateSQL(change schema.Change) (string, error) {
	switch c := change.(type) {
	// Schema-related changes - MySQL uses databases instead of schemas
	case schema.CreateSchemaChange:
		return my.generateCreateDatabase(c), nil
	case schema.DropSchemaChange:
		return my.generateDropDatabase(c), nil

	// Extension-related changes - Not supported in MySQL
	case schema.EnableExtensionChange:
		return "", fmt.Errorf("extensions not supported in MySQL")
	case schema.DisableExtensionChange:
		return "", fmt.Errorf("extensions not supported in MySQL")

	// Table-related changes
	case schema.CreateTableChange:
		return my.generateCreateTable(c), nil
	case schema.DropTableChange:
		return my.generateDropTable(c), nil

	// Column-related changes
	case schema.AddColumnChange:
		return my.generateAddColumn(c), nil
	case schema.DropColumnChange:
		return my.generateDropColumn(c), nil
	case schema.AlterColumnChange:
		return my.generateAlterColumn(c), nil

	// Primary key-related changes
	case schema.AddPrimaryKeyChange:
		return my.generateAddPrimaryKey(c), nil
	case schema.DropPrimaryKeyChange:
		return my.generateDropPrimaryKey(c), nil

	// Index-related changes
	case schema.AddIndexChange:
		return my.generateAddIndex(c), nil
	case schema.DropIndexChange:
		return my.generateDropIndex(c), nil

	// Foreign key-related changes
	case schema.AddForeignKeyChange:
		return my.generateAddForeignKey(c), nil
	case schema.DropForeignKeyChange:
		return my.generateDropForeignKey(c), nil

	// Sequence-related changes - Not directly supported in MySQL
	case schema.CreateSequenceChange:
		return "", fmt.Errorf("sequences not supported in MySQL, use AUTO_INCREMENT instead")
	case schema.AlterSequenceChange:
		return "", fmt.Errorf("sequences not supported in MySQL, use AUTO_INCREMENT instead")
	case schema.DropSequenceChange:
		return "", fmt.Errorf("sequences not supported in MySQL, use AUTO_INCREMENT instead")

	// Function-related changes
	case schema.CreateFunctionChange:
		return my.generateCreateFunction(c), nil
	case schema.AlterFunctionChange:
		return my.generateAlterFunction(c), nil
	case schema.DropFunctionChange:
		return my.generateDropFunction(c), nil

	// View-related changes
	case schema.CreateViewChange:
		return my.generateCreateView(c), nil
	case schema.AlterViewChange:
		return my.generateAlterView(c), nil
	case schema.DropViewChange:
		return my.generateDropView(c), nil

	// Trigger-related changes
	case schema.CreateTriggerChange:
		return my.generateCreateTrigger(c), nil
	case schema.AlterTriggerChange:
		// MySQL doesn't support directly altering triggers, so we drop and recreate
		return my.generateAlterTrigger(c), nil
	case schema.DropTriggerChange:
		return my.generateDropTrigger(c), nil

	default:
		return "", fmt.Errorf("unsupported change type: %T", change)
	}
}

// Helper function to quote MySQL identifiers
func quoteIdentifier(name string) string {
	return "`" + strings.Replace(name, "`", "``", -1) + "`"
}

// Helper function to quote MySQL string literals
func quoteLiteral(str string) string {
	return "'" + strings.Replace(str, "'", "''", -1) + "'"
}

// Schema-related SQL generation (MySQL uses databases instead of schemas)

func (my *MySQL) generateCreateDatabase(c schema.CreateSchemaChange) string {
	return fmt.Sprintf("CREATE DATABASE %s;", quoteIdentifier(c.SchemaName))
}

func (my *MySQL) generateDropDatabase(c schema.DropSchemaChange) string {
	return fmt.Sprintf("DROP DATABASE %s;", quoteIdentifier(c.SchemaName))
}

// Table-related SQL generation

func (my *MySQL) generateCreateTable(c schema.CreateTableChange) string {
	table := c.TableDef
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("CREATE TABLE %s (\n", quoteIdentifier(table.Name)))

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
		sb.WriteString(fmt.Sprintf(",\n  PRIMARY KEY (%s)",
			strings.Join(pkColumnsList, ", ")))
	}

	sb.WriteString("\n) ENGINE=InnoDB;")

	// Add comments for columns if present (MySQL syntax differs from PostgreSQL)
	for _, col := range table.Columns {
		if col.Comment != "" {
			sb.WriteString(fmt.Sprintf("\nALTER TABLE %s MODIFY COLUMN %s %s",
				quoteIdentifier(table.Name),
				quoteIdentifier(col.Name),
				col.TypeSQL()))

			if !col.Nullable {
				sb.WriteString(" NOT NULL")
			}

			if col.Default != "" {
				sb.WriteString(fmt.Sprintf(" DEFAULT %s", col.Default))
			}

			sb.WriteString(fmt.Sprintf(" COMMENT %s;", quoteLiteral(col.Comment)))
		}
	}

	return sb.String()
}

func (my *MySQL) generateDropTable(c schema.DropTableChange) string {
	return fmt.Sprintf("DROP TABLE %s;", quoteIdentifier(c.TableName))
}

// Column-related SQL generation

func (my *MySQL) generateAddColumn(c schema.AddColumnChange) string {
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

	if column.Comment != "" {
		sql += fmt.Sprintf(" COMMENT %s", quoteLiteral(column.Comment))
	}

	sql += ";"

	return sql
}

func (my *MySQL) generateDropColumn(c schema.DropColumnChange) string {
	return fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s;",
		quoteIdentifier(c.TableName),
		quoteIdentifier(c.ColumnName))
}

func (my *MySQL) generateAlterColumn(c schema.AlterColumnChange) string {
	column := c.Column
	sql := fmt.Sprintf("ALTER TABLE %s MODIFY COLUMN %s %s",
		quoteIdentifier(c.TableName),
		quoteIdentifier(column.Name),
		column.TypeSQL())

	if !column.Nullable {
		sql += " NOT NULL"
	}

	if column.Default != "" {
		sql += fmt.Sprintf(" DEFAULT %s", column.Default)
	}

	if column.Comment != "" {
		sql += fmt.Sprintf(" COMMENT %s", quoteLiteral(column.Comment))
	}

	sql += ";"
	return sql
}

// Primary key-related SQL generation

func (my *MySQL) generateAddPrimaryKey(c schema.AddPrimaryKeyChange) string {
	pk := c.PrimaryKey
	columns := make([]string, len(pk.Columns))
	for i, col := range pk.Columns {
		columns[i] = quoteIdentifier(col)
	}

	return fmt.Sprintf("ALTER TABLE %s ADD PRIMARY KEY (%s);",
		quoteIdentifier(c.TableName),
		strings.Join(columns, ", "))
}

func (my *MySQL) generateDropPrimaryKey(c schema.DropPrimaryKeyChange) string {
	return fmt.Sprintf("ALTER TABLE %s DROP PRIMARY KEY;",
		quoteIdentifier(c.TableName))
}

// Index-related SQL generation

func (my *MySQL) generateAddIndex(c schema.AddIndexChange) string {
	index := c.Index
	indexType := "INDEX"
	if index.Unique {
		indexType = "UNIQUE INDEX"
	}

	columns := make([]string, len(index.Columns))
	for i, col := range index.Columns {
		columns[i] = quoteIdentifier(col)
	}

	return fmt.Sprintf("CREATE %s %s ON %s (%s);",
		indexType,
		quoteIdentifier(index.Name),
		quoteIdentifier(c.TableName),
		strings.Join(columns, ", "))
}

func (my *MySQL) generateDropIndex(c schema.DropIndexChange) string {
	return fmt.Sprintf("DROP INDEX %s ON %s;",
		quoteIdentifier(c.IndexName),
		quoteIdentifier(c.TableName))
}

// Foreign key-related SQL generation

func (my *MySQL) generateAddForeignKey(c schema.AddForeignKeyChange) string {
	fk := c.ForeignKey
	columns := make([]string, len(fk.Columns))
	for i, col := range fk.Columns {
		columns[i] = quoteIdentifier(col)
	}

	refColumns := make([]string, len(fk.RefColumns))
	for i, col := range fk.RefColumns {
		refColumns[i] = quoteIdentifier(col)
	}

	sql := fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s (%s)",
		quoteIdentifier(c.TableName),
		quoteIdentifier(fk.Name),
		strings.Join(columns, ", "),
		quoteIdentifier(fk.RefTable),
		strings.Join(refColumns, ", "))

	if fk.OnDelete != "" {
		sql += fmt.Sprintf(" ON DELETE %s", fk.OnDelete)
	}

	if fk.OnUpdate != "" {
		sql += fmt.Sprintf(" ON UPDATE %s", fk.OnUpdate)
	}

	sql += ";"
	return sql
}

func (my *MySQL) generateDropForeignKey(c schema.DropForeignKeyChange) string {
	return fmt.Sprintf("ALTER TABLE %s DROP FOREIGN KEY %s;",
		quoteIdentifier(c.TableName),
		quoteIdentifier(c.FKName))
}

// Function-related SQL generation

func (my *MySQL) generateCreateFunction(c schema.CreateFunctionChange) string {
	fn := c.Function
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("CREATE FUNCTION %s(", quoteIdentifier(fn.Name)))

	// Function arguments
	args := make([]string, len(fn.Arguments))
	for i, arg := range fn.Arguments {
		argStr := ""
		if arg.Name != "" {
			argStr += arg.Name + " "
		}
		argStr += arg.Type
		args[i] = argStr
	}
	sb.WriteString(strings.Join(args, ", "))
	sb.WriteString(fmt.Sprintf(")\nRETURNS %s\n", fn.Returns))

	// Function attributes
	if fn.Volatility == "IMMUTABLE" {
		sb.WriteString("DETERMINISTIC\n")
	} else {
		sb.WriteString("NOT DETERMINISTIC\n")
	}

	// MySQL doesn't support all PostgreSQL function attributes

	// Function body
	sb.WriteString("BEGIN\n")
	sb.WriteString(fn.Body)
	sb.WriteString("\nEND;")

	return sb.String()
}

func (my *MySQL) generateAlterFunction(c schema.AlterFunctionChange) string {
	// MySQL doesn't support ALTER FUNCTION for changing the body, need to drop and recreate
	return fmt.Sprintf("-- MySQL doesn't support ALTER FUNCTION for changing the body.\n-- Drop and recreate the function instead:\n\nDROP FUNCTION IF EXISTS %s;\n%s",
		quoteIdentifier(c.Function.Name),
		my.generateCreateFunction(schema.CreateFunctionChange{Function: c.Function}))
}

func (my *MySQL) generateDropFunction(c schema.DropFunctionChange) string {
	return fmt.Sprintf("DROP FUNCTION IF EXISTS %s;",
		quoteIdentifier(c.FunctionName))
}

// View-related SQL generation

func (my *MySQL) generateCreateView(c schema.CreateViewChange) string {
	view := c.View
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("CREATE VIEW %s ", quoteIdentifier(view.Name)))

	if len(view.Columns) > 0 {
		columns := make([]string, len(view.Columns))
		for i, col := range view.Columns {
			columns[i] = quoteIdentifier(col)
		}
		sb.WriteString(fmt.Sprintf("(%s) ", strings.Join(columns, ", ")))
	}

	sb.WriteString("AS\n")
	sb.WriteString(view.Definition)
	sb.WriteString(";")

	return sb.String()
}

func (my *MySQL) generateAlterView(c schema.AlterViewChange) string {
	// MySQL uses the same syntax for CREATE OR REPLACE VIEW
	view := c.View
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("CREATE OR REPLACE VIEW %s ", quoteIdentifier(view.Name)))

	if len(view.Columns) > 0 {
		columns := make([]string, len(view.Columns))
		for i, col := range view.Columns {
			columns[i] = quoteIdentifier(col)
		}
		sb.WriteString(fmt.Sprintf("(%s) ", strings.Join(columns, ", ")))
	}

	sb.WriteString("AS\n")
	sb.WriteString(view.Definition)
	sb.WriteString(";")

	return sb.String()
}

func (my *MySQL) generateDropView(c schema.DropViewChange) string {
	return fmt.Sprintf("DROP VIEW IF EXISTS %s;", quoteIdentifier(c.ViewName))
}

// Trigger-related SQL generation

func (my *MySQL) generateCreateTrigger(c schema.CreateTriggerChange) string {
	trigger := c.Trigger
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("CREATE TRIGGER %s\n", quoteIdentifier(trigger.Name)))
	sb.WriteString(fmt.Sprintf("%s %s ON %s\n",
		trigger.Timing,
		strings.Join(trigger.Events, " OR "),
		quoteIdentifier(trigger.Table)))

	sb.WriteString(fmt.Sprintf("FOR EACH %s\n", trigger.ForEach))

	if trigger.When != "" {
		// MySQL doesn't support WHEN conditions in the same way as PostgreSQL
		sb.WriteString(fmt.Sprintf("-- WHEN condition not directly supported in MySQL: %s\n", trigger.When))
	}

	// Body
	sb.WriteString("BEGIN\n")
	// In MySQL, the body would be the direct SQL statements, not a function call
	sb.WriteString(fmt.Sprintf("  -- Call function: %s\n", trigger.Function))
	sb.WriteString("  -- Replace with actual trigger body\n")
	sb.WriteString("END;")

	return sb.String()
}

func (my *MySQL) generateAlterTrigger(c schema.AlterTriggerChange) string {
	// MySQL doesn't support ALTER TRIGGER, need to drop and recreate
	trigger := c.Trigger
	return fmt.Sprintf("-- MySQL doesn't support ALTER TRIGGER.\n-- Drop and recreate the trigger instead:\n\nDROP TRIGGER IF EXISTS %s;\n%s",
		quoteIdentifier(trigger.Name),
		my.generateCreateTrigger(schema.CreateTriggerChange{Trigger: trigger}))
}

func (my *MySQL) generateDropTrigger(c schema.DropTriggerChange) string {
	return fmt.Sprintf("DROP TRIGGER IF EXISTS %s;", quoteIdentifier(c.TriggerName))
}

// CreateTable generates SQL to create a table
func (my *MySQL) CreateTable(table *schema.Table) string {
	var b strings.Builder

	fmt.Fprintf(&b, "CREATE TABLE %s (\n", QuoteIdentifier(table.Name))

	// Add columns
	for i, column := range table.Columns {
		if i > 0 {
			b.WriteString(",\n")
		}
		b.WriteString("  ")
		b.WriteString(my.CreateColumn(column))
	}

	// Add primary key
	if table.PrimaryKey != nil && len(table.PrimaryKey.Columns) > 0 {
		b.WriteString(",\n  ")
		b.WriteString(my.CreatePrimaryKey(table.PrimaryKey))
	}

	// Add unique constraints and indexes
	for _, index := range table.Indexes {
		if index.Unique {
			b.WriteString(",\n  ")
			b.WriteString(my.CreateUniqueConstraint(index))
		}
	}

	// Add foreign keys
	for _, fk := range table.ForeignKeys {
		b.WriteString(",\n  ")
		b.WriteString(my.CreateForeignKey(fk))
	}

	b.WriteString("\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci")

	return b.String()
}

// CreateColumn generates SQL for a column definition
func (my *MySQL) CreateColumn(column *schema.Column) string {
	var b strings.Builder

	fmt.Fprintf(&b, "%s %s", QuoteIdentifier(column.Name), column.TypeSQL())

	if !column.Nullable {
		b.WriteString(" NOT NULL")
	}

	if column.Default != "" {
		fmt.Fprintf(&b, " DEFAULT %s", column.Default)
	}

	// if column.Identity == "ALWAYS" || column.Identity == "BY DEFAULT" {
	// 	b.WriteString(" AUTO_INCREMENT")
	// }

	return b.String()
}

// CreatePrimaryKey generates SQL for a primary key constraint
func (my *MySQL) CreatePrimaryKey(pk *schema.PrimaryKey) string {
	var columns []string
	for _, col := range pk.Columns {
		columns = append(columns, QuoteIdentifier(col))
	}

	return fmt.Sprintf("PRIMARY KEY (%s)", strings.Join(columns, ", "))
}

// CreateUniqueConstraint generates SQL for a unique constraint
func (my *MySQL) CreateUniqueConstraint(index *schema.Index) string {
	var columns []string
	for _, col := range index.Columns {
		columns = append(columns, QuoteIdentifier(col))
	}

	return fmt.Sprintf("UNIQUE KEY %s (%s)",
		QuoteIdentifier(index.Name),
		strings.Join(columns, ", "))
}

// CreateIndex generates SQL to create an index
func (my *MySQL) CreateIndex(index *schema.Index, tableName string) string {
	var b strings.Builder

	if index.Unique {
		fmt.Fprintf(&b, "CREATE UNIQUE INDEX %s ON %s (",
			QuoteIdentifier(index.Name),
			QuoteIdentifier(tableName))
	} else {
		fmt.Fprintf(&b, "CREATE INDEX %s ON %s (",
			QuoteIdentifier(index.Name),
			QuoteIdentifier(tableName))
	}

	for i, col := range index.Columns {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(QuoteIdentifier(col))
	}
	b.WriteString(")")

	return b.String()
}

// CreateForeignKey generates SQL for a foreign key constraint
func (my *MySQL) CreateForeignKey(fk *schema.ForeignKey) string {
	var srcColumns, refColumns []string

	for _, col := range fk.Columns {
		srcColumns = append(srcColumns, QuoteIdentifier(col))
	}

	for _, col := range fk.RefColumns {
		refColumns = append(refColumns, QuoteIdentifier(col))
	}

	var b strings.Builder
	fmt.Fprintf(&b, "CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s (%s)",
		QuoteIdentifier(fk.Name),
		strings.Join(srcColumns, ", "),
		QuoteIdentifier(fk.RefTable),
		strings.Join(refColumns, ", "))

	if fk.OnDelete != "" {
		fmt.Fprintf(&b, " ON DELETE %s", fk.OnDelete)
	}

	if fk.OnUpdate != "" {
		fmt.Fprintf(&b, " ON UPDATE %s", fk.OnUpdate)
	}

	return b.String()
}

// CreateView generates SQL to create a view
func (my *MySQL) CreateView(view *schema.View) string {
	return fmt.Sprintf("CREATE VIEW %s AS %s",
		QuoteIdentifier(view.Name),
		view.Definition)
}

// CreateFunction generates SQL to create a function
func (my *MySQL) CreateFunction(function *schema.Function) string {
	var b strings.Builder

	fmt.Fprintf(&b, "CREATE FUNCTION %s(", QuoteIdentifier(function.Name))

	// Add arguments
	for i, arg := range function.Arguments {
		if i > 0 {
			b.WriteString(", ")
		}

		if arg.Name != "" {
			fmt.Fprintf(&b, "%s %s", arg.Name, arg.Type)
		} else {
			b.WriteString(arg.Type)
		}
	}

	b.WriteString(")")

	// Add return type
	fmt.Fprintf(&b, " RETURNS %s\n", function.Returns)

	// Add deterministic/not deterministic
	if function.Volatility == "IMMUTABLE" {
		b.WriteString("DETERMINISTIC\n")
	} else {
		b.WriteString("NOT DETERMINISTIC\n")
	}

	// Add function body
	fmt.Fprintf(&b, "BEGIN\n%s\nEND", function.Body)

	return b.String()
}

// CreateTrigger generates SQL to create a trigger
func (my *MySQL) CreateTrigger(trigger *schema.Trigger) string {
	var b strings.Builder

	fmt.Fprintf(&b, "CREATE TRIGGER %s\n", QuoteIdentifier(trigger.Name))
	fmt.Fprintf(&b, "%s %s ON %s\n",
		trigger.Timing,
		trigger.Events[0], // MySQL triggers have one event
		QuoteIdentifier(trigger.Table))
	fmt.Fprintf(&b, "FOR EACH ROW\n")
	b.WriteString(trigger.Function) // In MySQL, this contains the actual trigger code

	return b.String()
}

// QuoteIdentifier quotes an identifier (table, column name, etc.) for MySQL
func QuoteIdentifier(s string) string {
	return "`" + strings.Replace(s, "`", "``", -1) + "`"
}
