package postgresql

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/swiftcarrot/dbx/internal/testutil"
	"github.com/swiftcarrot/dbx/schema"
)

func TestCreateSchema(t *testing.T) {
	pg := New()
	createSchema := schema.CreateSchemaChange{
		SchemaName: "test_schema",
	}
	sql, err := pg.GenerateSQL(createSchema)
	require.NoError(t, err)
	require.Equal(t, `CREATE SCHEMA "test_schema";`, sql)
}

func TestDropSchema(t *testing.T) {
	pg := New()
	dropSchema := schema.DropSchemaChange{
		SchemaName: "test_schema",
	}
	sql, err := pg.GenerateSQL(dropSchema)
	require.NoError(t, err)
	require.Equal(t, `DROP SCHEMA "test_schema";`, sql)
}

func TestEnableExtension(t *testing.T) {
	pg := New()
	enableExt := schema.EnableExtensionChange{
		Extension: "uuid-ossp",
	}
	sql, err := pg.GenerateSQL(enableExt)
	require.NoError(t, err)
	require.Equal(t, `CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`, sql)
}

func TestDisableExtension(t *testing.T) {
	pg := New()
	disableExt := schema.DisableExtensionChange{
		Extension: "uuid-ossp",
	}
	sql, err := pg.GenerateSQL(disableExt)
	require.NoError(t, err)
	require.Equal(t, `DROP EXTENSION IF EXISTS "uuid-ossp";`, sql)
}

func TestCreateTable(t *testing.T) {
	pg := New()

	// Test simple table creation
	table := &schema.Table{
		Name: "users",
		Columns: []*schema.Column{
			{Name: "id", Type: &schema.IntegerType{}, Nullable: false, AutoIncrement: true},
			{Name: "name", Type: &schema.VarcharType{Length: 100}, Nullable: false},
			{Name: "email", Type: &schema.VarcharType{Length: 255}, Nullable: false},
			{Name: "bio", Type: &schema.TextType{}, Nullable: true, Comment: "User biography"},
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
	sql, err := pg.GenerateSQL(createTable)
	require.NoError(t, err)
	expected := `CREATE TABLE "users" (
  "id" serial NOT NULL,
  "name" varchar(100) NOT NULL,
  "email" varchar(255) NOT NULL,
  "bio" text,
  "created_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT "users_pkey" PRIMARY KEY ("id")
);
COMMENT ON COLUMN "users"."bio" IS 'User biography';`
	require.Equal(t, testutil.FormatSQL(expected), testutil.FormatSQL(sql))

	// Test schema-qualified table
	schemaTable := &schema.Table{
		Schema: "test_schema",
		Name:   "users",
		Columns: []*schema.Column{
			{Name: "id", Type: &SerialType{}, Nullable: false},
		},
	}
	createSchemaTable := schema.CreateTableChange{
		TableDef: schemaTable,
	}
	sql, err = pg.GenerateSQL(createSchemaTable)
	require.NoError(t, err)
	require.Contains(t, sql, `CREATE TABLE "test_schema"."users"`)
}

func TestDropTable(t *testing.T) {
	pg := New()

	// Test simple table drop
	dropTable := schema.DropTableChange{
		TableName: "users",
	}
	sql, err := pg.GenerateSQL(dropTable)
	require.NoError(t, err)
	require.Equal(t, `DROP TABLE "users";`, sql)

	// Test schema-qualified table drop
	dropSchemaTable := schema.DropTableChange{
		SchemaName: "test_schema",
		TableName:  "users",
	}
	sql, err = pg.GenerateSQL(dropSchemaTable)
	require.NoError(t, err)
	require.Equal(t, `DROP TABLE "test_schema"."users";`, sql)
}

func TestAddColumn(t *testing.T) {
	pg := New()
	column := &schema.Column{
		Name:     "email",
		Type:     &schema.VarcharType{Length: 255},
		Nullable: false,
		Default:  "'user@example.com'",
		Comment:  "User email address",
	}
	addColumn := schema.AddColumnChange{
		TableName: "users",
		Column:    column,
	}
	sql, err := pg.GenerateSQL(addColumn)
	require.NoError(t, err)
	expected := `ALTER TABLE "users" ADD COLUMN "email" varchar(255) NOT NULL DEFAULT 'user@example.com';
COMMENT ON COLUMN "users"."email" IS 'User email address';`
	require.Equal(t, testutil.FormatSQL(expected), testutil.FormatSQL(sql))
}

func TestDropColumn(t *testing.T) {
	pg := New()
	dropColumn := schema.DropColumnChange{
		TableName:  "users",
		ColumnName: "email",
	}
	sql, err := pg.GenerateSQL(dropColumn)
	require.NoError(t, err)
	require.Equal(t, `ALTER TABLE "users" DROP COLUMN "email";`, sql)
}

func TestAlterColumn(t *testing.T) {
	pg := New()
	alterColumn := schema.AlterColumnChange{
		TableName: "users",
		Column: &schema.Column{
			Name:     "email",
			Type:     &schema.VarcharType{Length: 100},
			Nullable: true,
			Default:  "",
			Comment:  "Updated comment",
		},
	}
	sql, err := pg.GenerateSQL(alterColumn)
	require.NoError(t, err)
	expected := `ALTER TABLE "users" ALTER COLUMN "email" TYPE varchar(100);
ALTER TABLE "users" ALTER COLUMN "email" DROP NOT NULL;
ALTER TABLE "users" ALTER COLUMN "email" DROP DEFAULT;
COMMENT ON COLUMN "users"."email" IS 'Updated comment';`
	require.Equal(t, testutil.FormatSQL(expected), testutil.FormatSQL(sql))
}

func TestAddPrimaryKey(t *testing.T) {
	pg := New()

	// Test simple primary key
	addPK := schema.AddPrimaryKeyChange{
		TableName: "users",
		PrimaryKey: &schema.PrimaryKey{
			Name:    "users_pkey",
			Columns: []string{"id"},
		},
	}
	sql, err := pg.GenerateSQL(addPK)
	require.NoError(t, err)
	require.Equal(t, `ALTER TABLE "users" ADD CONSTRAINT "users_pkey" PRIMARY KEY ("id");`, sql)

	// Test composite primary key
	addCompositePK := schema.AddPrimaryKeyChange{
		TableName: "order_items",
		PrimaryKey: &schema.PrimaryKey{
			Name:    "order_items_pkey",
			Columns: []string{"order_id", "product_id"},
		},
	}
	sql, err = pg.GenerateSQL(addCompositePK)
	require.NoError(t, err)
	require.Equal(t, `ALTER TABLE "order_items" ADD CONSTRAINT "order_items_pkey" PRIMARY KEY ("order_id", "product_id");`, sql)
}

func TestDropPrimaryKey(t *testing.T) {
	pg := New()
	dropPK := schema.DropPrimaryKeyChange{
		TableName: "users",
		PKName:    "users_pkey",
	}
	sql, err := pg.GenerateSQL(dropPK)
	require.NoError(t, err)
	require.Equal(t, `ALTER TABLE "users" DROP CONSTRAINT "users_pkey";`, sql)
}

func TestAddIndex(t *testing.T) {
	pg := New()

	// Test simple index
	addIdx := schema.AddIndexChange{
		TableName: "users",
		Index: &schema.Index{
			Name:    "idx_users_email",
			Columns: []string{"email"},
			Unique:  false,
		},
	}
	sql, err := pg.GenerateSQL(addIdx)
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
	sql, err = pg.GenerateSQL(addUniqueIdx)
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
	sql, err = pg.GenerateSQL(addMultiIdx)
	require.NoError(t, err)
	require.Equal(t, `CREATE INDEX "idx_users_name_email" ON "users" ("name", "email");`, sql)
}

func TestDropIndex(t *testing.T) {
	pg := New()
	dropIdx := schema.DropIndexChange{
		IndexName: "idx_users_email",
	}
	sql, err := pg.GenerateSQL(dropIdx)
	require.NoError(t, err)
	require.Equal(t, `DROP INDEX "idx_users_email";`, sql)
}

func TestAddForeignKey(t *testing.T) {
	pg := New()

	// Test simple foreign key
	addFK := schema.AddForeignKeyChange{
		TableName: "posts",
		ForeignKey: &schema.ForeignKey{
			Name:       "fk_posts_user",
			Columns:    []string{"user_id"},
			RefTable:   "users",
			RefColumns: []string{"id"},
			OnDelete:   "CASCADE",
		},
	}
	sql, err := pg.GenerateSQL(addFK)
	require.NoError(t, err)
	require.Equal(t, `ALTER TABLE "posts" ADD CONSTRAINT "fk_posts_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE;`, sql)

	// Test foreign key with ON UPDATE
	addFKUpdate := schema.AddForeignKeyChange{
		TableName: "posts",
		ForeignKey: &schema.ForeignKey{
			Name:       "fk_posts_category",
			Columns:    []string{"category_id"},
			RefTable:   "categories",
			RefColumns: []string{"id"},
			OnUpdate:   "CASCADE",
		},
	}
	sql, err = pg.GenerateSQL(addFKUpdate)
	require.NoError(t, err)
	require.Equal(t, `ALTER TABLE "posts" ADD CONSTRAINT "fk_posts_category" FOREIGN KEY ("category_id") REFERENCES "categories" ("id") ON UPDATE CASCADE;`, sql)

	// Test composite foreign key
	addCompositeFK := schema.AddForeignKeyChange{
		TableName: "order_items",
		ForeignKey: &schema.ForeignKey{
			Name:       "fk_order_items",
			Columns:    []string{"order_id", "product_id"},
			RefTable:   "products",
			RefColumns: []string{"order_id", "id"},
		},
	}
	sql, err = pg.GenerateSQL(addCompositeFK)
	require.NoError(t, err)
	require.Equal(t, `ALTER TABLE "order_items" ADD CONSTRAINT "fk_order_items" FOREIGN KEY ("order_id", "product_id") REFERENCES "products" ("order_id", "id");`, sql)
}

func TestDropForeignKey(t *testing.T) {
	pg := New()
	dropFK := schema.DropForeignKeyChange{
		TableName: "posts",
		FKName:    "fk_posts_user",
	}
	sql, err := pg.GenerateSQL(dropFK)
	require.NoError(t, err)
	require.Equal(t, `ALTER TABLE "posts" DROP CONSTRAINT "fk_posts_user";`, sql)
}

func TestCreateSequence(t *testing.T) {
	pg := New()

	// Test sequence with custom settings
	createSeq := schema.CreateSequenceChange{
		Sequence: &schema.Sequence{
			Name:      "order_id_seq",
			Start:     1000,
			Increment: 10,
			MinValue:  1000,
			MaxValue:  2147483647,
			Cache:     10,
			Cycle:     true,
		},
	}
	sql, err := pg.GenerateSQL(createSeq)
	require.NoError(t, err)
	expected := `CREATE SEQUENCE "order_id_seq" INCREMENT BY 10 MINVALUE 1000 MAXVALUE 2147483647 START WITH 1000 CACHE 10 CYCLE;`
	require.Equal(t, expected, sql)

	// Test sequence with default values
	createDefaultSeq := schema.CreateSequenceChange{
		Sequence: &schema.Sequence{
			Name: "user_id_seq",
		},
	}
	sql, err = pg.GenerateSQL(createDefaultSeq)
	require.NoError(t, err)
	require.Equal(t, `CREATE SEQUENCE "user_id_seq" INCREMENT BY 0 MINVALUE 0 MAXVALUE 0 START WITH 0 CACHE 0;`, sql)

	// Test schema-qualified sequence
	schemaSeq := schema.CreateSequenceChange{
		Sequence: &schema.Sequence{
			Schema: "test_schema",
			Name:   "order_id_seq",
		},
	}
	sql, err = pg.GenerateSQL(schemaSeq)
	require.NoError(t, err)
	require.Contains(t, sql, `CREATE SEQUENCE "test_schema"."order_id_seq"`)
}

func TestAlterSequence(t *testing.T) {
	pg := New()
	alterSeq := schema.AlterSequenceChange{
		Sequence: &schema.Sequence{
			Name:      "order_id_seq",
			Increment: 5,
			MinValue:  0,
			MaxValue:  1000000,
			Cache:     20,
			Cycle:     false,
		},
	}
	sql, err := pg.GenerateSQL(alterSeq)
	require.NoError(t, err)
	expected := `ALTER SEQUENCE "order_id_seq" INCREMENT BY 5 MINVALUE 0 MAXVALUE 1000000 CACHE 20 NO CYCLE;`
	require.Equal(t, expected, sql)
}

func TestDropSequence(t *testing.T) {
	pg := New()
	dropSeq := schema.DropSequenceChange{
		SequenceName: "order_id_seq",
	}
	sql, err := pg.GenerateSQL(dropSeq)
	require.NoError(t, err)
	require.Equal(t, `DROP SEQUENCE "order_id_seq";`, sql)
}

func TestCreateFunction(t *testing.T) {
	pg := New()
	createFn := schema.CreateFunctionChange{
		Function: &schema.Function{
			Name: "add_numbers",
			Arguments: []schema.FunctionArg{
				{Name: "a", Type: "integer"},
				{Name: "b", Type: "integer", Default: "0"},
			},
			Returns:  "integer",
			Language: "plpgsql",
			Body: `
BEGIN
  RETURN a + b;
END;
`,
			Volatility: "IMMUTABLE",
			Strict:     true,
			Security:   "INVOKER",
			Cost:       100,
		},
	}
	sql, err := pg.GenerateSQL(createFn)
	require.NoError(t, err)
	expected := `CREATE FUNCTION "add_numbers"(a integer, b integer DEFAULT 0) RETURNS integer AS $$
BEGIN
  RETURN a + b;
END;
$$ LANGUAGE plpgsql IMMUTABLE STRICT COST 100;`
	require.Equal(t, testutil.FormatSQL(expected), testutil.FormatSQL(sql))
}

func TestAlterFunction(t *testing.T) {
	pg := New()
	alterFn := schema.AlterFunctionChange{
		Function: &schema.Function{
			Name: "add_numbers",
			Arguments: []schema.FunctionArg{
				{Name: "a", Type: "integer"},
				{Name: "b", Type: "integer", Default: "0"},
			},
			Returns:  "integer",
			Language: "plpgsql",
			Body: `
BEGIN
  RETURN a + b + 1; -- Updated
END;
`,
			Volatility: "STABLE",
			Strict:     false,
			Security:   "DEFINER",
			Cost:       200,
		},
	}
	sql, err := pg.GenerateSQL(alterFn)
	require.NoError(t, err)
	expected := `CREATE OR REPLACE FUNCTION "add_numbers"(a integer, b integer DEFAULT 0) RETURNS integer AS $$
BEGIN
  RETURN a + b + 1; -- Updated
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER COST 200;`
	require.Equal(t, testutil.FormatSQL(expected), testutil.FormatSQL(sql))
}

func TestDropFunction(t *testing.T) {
	pg := New()

	// Test dropping function with arguments
	dropFn := schema.DropFunctionChange{
		FunctionName: "add_numbers",
		FunctionArgs: []schema.FunctionArg{
			{Type: "integer"},
			{Type: "integer"},
		},
	}
	sql, err := pg.GenerateSQL(dropFn)
	require.NoError(t, err)
	require.Equal(t, `DROP FUNCTION "add_numbers"(integer, integer);`, sql)

	// Test dropping function without arguments
	dropFnNoArgs := schema.DropFunctionChange{
		FunctionName: "current_timestamp",
	}
	sql, err = pg.GenerateSQL(dropFnNoArgs)
	require.NoError(t, err)
	require.Equal(t, `DROP FUNCTION "current_timestamp";`, sql)
}

func TestCreateView(t *testing.T) {
	pg := New()

	// Test view with columns and options
	createView := schema.CreateViewChange{
		View: &schema.View{
			Name:       "active_users",
			Columns:    []string{"id", "name", "email"},
			Definition: "SELECT id, name, email FROM users WHERE active = true",
			Options:    []string{"check_option=local"},
		},
	}
	sql, err := pg.GenerateSQL(createView)
	require.NoError(t, err)
	expected := `CREATE VIEW "active_users" ("id", "name", "email") WITH (check_option=local) AS SELECT id, name, email FROM users WHERE active = true;`
	require.Equal(t, testutil.FormatSQL(expected), testutil.FormatSQL(sql))

	// Test simple view
	simpleView := schema.CreateViewChange{
		View: &schema.View{
			Name:       "all_users",
			Definition: "SELECT * FROM users",
		},
	}
	sql, err = pg.GenerateSQL(simpleView)
	require.NoError(t, err)
	require.Equal(t, `CREATE VIEW "all_users" AS SELECT * FROM users;`, sql)
}

func TestAlterView(t *testing.T) {
	pg := New()
	alterView := schema.AlterViewChange{
		View: &schema.View{
			Name:       "active_users",
			Definition: "SELECT id, name, email FROM users WHERE active = true AND verified = true",
		},
	}
	sql, err := pg.GenerateSQL(alterView)
	require.NoError(t, err)
	expected := `CREATE VIEW "active_users" AS SELECT id, name, email FROM users WHERE active = true AND verified = true;`
	require.Equal(t, testutil.FormatSQL(expected), testutil.FormatSQL(sql))
}

func TestDropView(t *testing.T) {
	pg := New()
	dropView := schema.DropViewChange{
		ViewName: "active_users",
	}
	sql, err := pg.GenerateSQL(dropView)
	require.NoError(t, err)
	require.Equal(t, `DROP VIEW "active_users";`, sql)
}

func TestCreateTrigger(t *testing.T) {
	pg := New()

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
	sql, err := pg.GenerateSQL(createTrigger)
	require.NoError(t, err)
	expected := `CREATE TRIGGER "update_timestamp"
BEFORE UPDATE ON "users"
FOR EACH ROW
EXECUTE FUNCTION update_modified_column();`
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
	sql, err = pg.GenerateSQL(condTrigger)
	require.NoError(t, err)
	expected = `CREATE TRIGGER "log_changes"
AFTER INSERT OR UPDATE OR DELETE ON "products"
FOR EACH ROW
WHEN (OLD.price IS DISTINCT FROM NEW.price)
EXECUTE FUNCTION log_price_change('products', 'price');`
	require.Equal(t, testutil.FormatSQL(expected), testutil.FormatSQL(sql))
}

func TestAlterTrigger(t *testing.T) {
	pg := New()
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
	sql, err := pg.GenerateSQL(alterTrigger)
	require.NoError(t, err)
	expected := `DROP TRIGGER "update_timestamp" ON "users";
CREATE TRIGGER "update_timestamp"
BEFORE UPDATE OR INSERT ON "users"
FOR EACH ROW
EXECUTE FUNCTION update_modified_column();`
	require.Equal(t, testutil.FormatSQL(expected), testutil.FormatSQL(sql))
}

func TestDropTrigger(t *testing.T) {
	pg := New()
	dropTrigger := schema.DropTriggerChange{
		TriggerName:  "update_timestamp",
		TriggerTable: "users",
	}
	sql, err := pg.GenerateSQL(dropTrigger)
	require.NoError(t, err)
	require.Equal(t, `DROP TRIGGER "update_timestamp" ON "users";`, sql)
}

func TestQuoteIdentifier(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"users", `"users"`},
		{"user table", `"user table"`},
		{`users"table`, `"users""table"`},
		{"public.users", `"public"."users"`},
		{"test_schema.users", `"test_schema"."users"`},
		{`"already_quoted"`, `"already_quoted"`},
		{`"test_schema"."users"`, `"test_schema"."users"`},
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
