package sqlite

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/swiftcarrot/dbx/internal/testutil"
	"github.com/swiftcarrot/dbx/schema"
)

func TestCreateTable(t *testing.T) {
	sqlite := New()

	// Test simple table creation
	table := &schema.Table{
		Name: "users",
		Columns: []*schema.Column{
			{Name: "id", Type: &schema.IntegerType{}, Nullable: false},
			{Name: "name", Type: &schema.TextType{}, Nullable: false},
			{Name: "email", Type: &schema.TextType{}, Nullable: false},
			{Name: "bio", Type: &schema.TextType{}, Nullable: true},
			{Name: "created_at", Type: &schema.TimestampType{}, Nullable: false, Default: "CURRENT_TIMESTAMP"},
		},
		PrimaryKey: &schema.PrimaryKey{
			Name:    "users_pkey",
			Columns: []string{"id"},
		},
	}
	createTable := schema.CreateTableChange{
		TableDef: table,
	}
	sql, err := sqlite.GenerateSQL(createTable)
	require.NoError(t, err)
	expected := `CREATE TABLE "users" (
  "id" INTEGER NOT NULL,
  "name" TEXT NOT NULL,
  "email" TEXT NOT NULL,
  "bio" TEXT,
  "created_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT "users_pkey" PRIMARY KEY ("id")
);`
	require.Equal(t, testutil.FormatSQL(expected), testutil.FormatSQL(sql))
}

func TestDropTable(t *testing.T) {
	sqlite := New()
	dropTable := schema.DropTableChange{
		TableName: "users",
	}
	sql, err := sqlite.GenerateSQL(dropTable)
	require.NoError(t, err)
	require.Equal(t, `DROP TABLE "users";`, sql)
}

func TestAddColumn(t *testing.T) {
	sqlite := New()
	column := &schema.Column{
		Name:     "email",
		Type:     &schema.TextType{},
		Nullable: false,
		Default:  "'user@example.com'",
	}
	addColumn := schema.AddColumnChange{
		TableName: "users",
		Column:    column,
	}
	sql, err := sqlite.GenerateSQL(addColumn)
	require.NoError(t, err)
	expected := `ALTER TABLE "users" ADD COLUMN "email" TEXT NOT NULL DEFAULT 'user@example.com';`
	require.Equal(t, testutil.FormatSQL(expected), testutil.FormatSQL(sql))
}

func TestAddIndex(t *testing.T) {
	sqlite := New()

	// Test simple index
	addIdx := schema.AddIndexChange{
		TableName: "users",
		Index: &schema.Index{
			Name:    "idx_users_email",
			Columns: []string{"email"},
			Unique:  false,
		},
	}
	sql, err := sqlite.GenerateSQL(addIdx)
	require.NoError(t, err)
	require.Equal(t, `CREATE INDEX "idx_users_email" ON "users" ("email");`, sql)

	// Test unique index
	addUniqueIdx := schema.AddIndexChange{
		TableName: "users",
		Index: &schema.Index{
			Name:    "idx_users_email_unique",
			Columns: []string{"email"},
			Unique:  true,
		},
	}
	sql, err = sqlite.GenerateSQL(addUniqueIdx)
	require.NoError(t, err)
	require.Equal(t, `CREATE UNIQUE INDEX "idx_users_email_unique" ON "users" ("email");`, sql)

	// Test multi-column index
	addMultiIdx := schema.AddIndexChange{
		TableName: "users",
		Index: &schema.Index{
			Name:    "idx_users_name_email",
			Columns: []string{"name", "email"},
		},
	}
	sql, err = sqlite.GenerateSQL(addMultiIdx)
	require.NoError(t, err)
	require.Equal(t, `CREATE INDEX "idx_users_name_email" ON "users" ("name", "email");`, sql)
}

func TestDropIndex(t *testing.T) {
	sqlite := New()
	dropIdx := schema.DropIndexChange{
		IndexName: "idx_users_email",
	}
	sql, err := sqlite.GenerateSQL(dropIdx)
	require.NoError(t, err)
	require.Equal(t, `DROP INDEX "idx_users_email";`, sql)
}

func TestCreateView(t *testing.T) {
	sqlite := New()

	// Test view with columns
	createView := schema.CreateViewChange{
		View: &schema.View{
			Name:       "active_users",
			Columns:    []string{"id", "name", "email"},
			Definition: "SELECT id, name, email FROM users WHERE active = 1",
		},
	}
	sql, err := sqlite.GenerateSQL(createView)
	require.NoError(t, err)
	expected := `CREATE VIEW "active_users" ("id", "name", "email") AS SELECT id, name, email FROM users WHERE active = 1;`
	require.Equal(t, testutil.FormatSQL(expected), testutil.FormatSQL(sql))

	// Test simple view
	simpleView := schema.CreateViewChange{
		View: &schema.View{
			Name:       "all_users",
			Definition: "SELECT * FROM users",
		},
	}
	sql, err = sqlite.GenerateSQL(simpleView)
	require.NoError(t, err)
	require.Equal(t, `CREATE VIEW "all_users" AS SELECT * FROM users;`, sql)
}

func TestAlterView(t *testing.T) {
	sqlite := New()
	alterView := schema.AlterViewChange{
		View: &schema.View{
			Name:       "active_users",
			Definition: "SELECT id, name, email FROM users WHERE active = 1 AND verified = 1",
		},
	}
	sql, err := sqlite.GenerateSQL(alterView)
	require.NoError(t, err)
	expected := `DROP VIEW IF EXISTS "active_users";
CREATE VIEW "active_users" AS SELECT id, name, email FROM users WHERE active = 1 AND verified = 1;`
	require.Equal(t, testutil.FormatSQL(expected), testutil.FormatSQL(sql))
}

func TestDropView(t *testing.T) {
	sqlite := New()
	dropView := schema.DropViewChange{
		ViewName: "active_users",
	}
	sql, err := sqlite.GenerateSQL(dropView)
	require.NoError(t, err)
	require.Equal(t, `DROP VIEW "active_users";`, sql)
}

func TestCreateTrigger(t *testing.T) {
	sqlite := New()

	// Test simple trigger
	createTrigger := schema.CreateTriggerChange{
		Trigger: &schema.Trigger{
			Name:      "update_timestamp",
			Table:     "users",
			Events:    []string{"UPDATE"},
			Timing:    "BEFORE",
			ForEach:   "ROW",
			Function:  "update_modified_column",
			Arguments: []string{},
		},
	}
	sql, err := sqlite.GenerateSQL(createTrigger)
	require.NoError(t, err)
	expected := `CREATE TRIGGER "update_timestamp"
BEFORE UPDATE ON "users"
FOR EACH ROW
BEGIN
  SELECT update_modified_column;
END;`
	require.Equal(t, testutil.FormatSQL(expected), testutil.FormatSQL(sql))

	// Test trigger with condition and arguments
	condTrigger := schema.CreateTriggerChange{
		Trigger: &schema.Trigger{
			Name:     "log_changes",
			Table:    "products",
			Events:   []string{"INSERT", "UPDATE", "DELETE"},
			Timing:   "AFTER",
			ForEach:  "ROW",
			When:     "OLD.price IS DISTINCT FROM NEW.price",
			Function: "log_price_change",
			Arguments: []string{
				"'products'",
				"'price'",
			},
		},
	}
	sql, err = sqlite.GenerateSQL(condTrigger)
	require.NoError(t, err)
	expected = `CREATE TRIGGER "log_changes"
AFTER INSERT OR UPDATE OR DELETE ON "products"
FOR EACH ROW
WHEN (OLD.price IS DISTINCT FROM NEW.price)
BEGIN
  SELECT log_price_change('products', 'price');
END;`
	require.Equal(t, testutil.FormatSQL(expected), testutil.FormatSQL(sql))
}

func TestAlterTrigger(t *testing.T) {
	sqlite := New()
	alterTrigger := schema.AlterTriggerChange{
		Trigger: &schema.Trigger{
			Name:     "update_timestamp",
			Table:    "users",
			Events:   []string{"UPDATE", "INSERT"},
			Timing:   "BEFORE",
			ForEach:  "ROW",
			Function: "update_modified_column",
		},
	}
	sql, err := sqlite.GenerateSQL(alterTrigger)
	require.NoError(t, err)
	expected := `DROP TRIGGER IF EXISTS "update_timestamp";
CREATE TRIGGER "update_timestamp"
BEFORE UPDATE OR INSERT ON "users"
FOR EACH ROW
BEGIN
  SELECT update_modified_column;
END;`
	require.Equal(t, testutil.FormatSQL(expected), testutil.FormatSQL(sql))
}

func TestDropTrigger(t *testing.T) {
	sqlite := New()
	dropTrigger := schema.DropTriggerChange{
		TriggerName: "update_timestamp",
	}
	sql, err := sqlite.GenerateSQL(dropTrigger)
	require.NoError(t, err)
	require.Equal(t, `DROP TRIGGER "update_timestamp";`, sql)
}

func TestUnsupportedOperations(t *testing.T) {
	sqlite := New()

	// Test dropping column (unsupported in SQLite)
	dropColumn := schema.DropColumnChange{
		TableName:  "users",
		ColumnName: "email",
	}
	_, err := sqlite.GenerateSQL(dropColumn)
	require.Error(t, err)
	require.Contains(t, err.Error(), "SQLite does not support DROP COLUMN directly")

	// Test altering column (unsupported in SQLite)
	alterColumn := schema.AlterColumnChange{
		TableName: "users",
		Column: &schema.Column{
			Name:     "email",
			Type:     &schema.TextType{},
			Nullable: true,
		},
	}
	_, err = sqlite.GenerateSQL(alterColumn)
	require.Error(t, err)
	require.Contains(t, err.Error(), "SQLite does not support ALTER COLUMN directly")

	// Test adding primary key (unsupported in SQLite)
	addPK := schema.AddPrimaryKeyChange{
		TableName: "users",
		PrimaryKey: &schema.PrimaryKey{
			Name:    "users_pkey",
			Columns: []string{"id"},
		},
	}
	_, err = sqlite.GenerateSQL(addPK)
	require.Error(t, err)
	require.Contains(t, err.Error(), "SQLite does not support adding primary keys")

	// Test adding foreign key (unsupported in SQLite)
	addFK := schema.AddForeignKeyChange{
		TableName: "orders",
		ForeignKey: &schema.ForeignKey{
			Name:       "fk_orders_users",
			Columns:    []string{"user_id"},
			RefTable:   "users",
			RefColumns: []string{"id"},
		},
	}
	_, err = sqlite.GenerateSQL(addFK)
	require.Error(t, err)
	require.Contains(t, err.Error(), "SQLite does not support adding foreign keys")
}

func TestQuoteIdentifier(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"users", `"users"`},
		{"user table", `"user table"`},
	}

	for _, test := range tests {
		require.Equal(t, test.expected, quoteIdentifier(test.input))
	}
}

func TestQuoteLiteral(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"test", `'test'`},
		{"test's string", `'test''s string'`},
		{"O'Reilly", `'O''Reilly'`},
		{`'already_quoted'`, `'already_quoted'`},
	}

	for _, test := range tests {
		require.Equal(t, test.expected, quoteLiteral(test.input))
	}
}
