package mysql

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/swiftcarrot/dbx/internal/testutil"
	"github.com/swiftcarrot/dbx/schema"
)

func TestCreateDatabase(t *testing.T) {
	my := New()
	createSchema := schema.CreateSchemaChange{
		SchemaName: "test_db",
	}
	sql, err := my.GenerateSQL(createSchema)
	require.NoError(t, err)
	require.Equal(t, "CREATE DATABASE `test_db`;", sql)
}

func TestDropDatabase(t *testing.T) {
	my := New()
	dropSchema := schema.DropSchemaChange{
		SchemaName: "test_db",
	}
	sql, err := my.GenerateSQL(dropSchema)
	require.NoError(t, err)
	require.Equal(t, "DROP DATABASE `test_db`;", sql)
}

func TestExtensionNotSupported(t *testing.T) {
	my := New()

	// Test enabling extension (not supported in MySQL)
	enableExt := schema.EnableExtensionChange{
		Extension: "uuid-ossp",
	}
	_, err := my.GenerateSQL(enableExt)
	require.Error(t, err)

	// Test disabling extension (not supported in MySQL)
	disableExt := schema.DisableExtensionChange{
		Extension: "uuid-ossp",
	}
	_, err = my.GenerateSQL(disableExt)
	require.Error(t, err)
}

func TestSequenceNotSupported(t *testing.T) {
	my := New()

	// Test creating sequence (not directly supported in MySQL)
	createSeq := schema.CreateSequenceChange{
		Sequence: &schema.Sequence{
			Name: "order_seq",
		},
	}
	_, err := my.GenerateSQL(createSeq)
	require.Error(t, err)

	// Test altering sequence (not directly supported in MySQL)
	alterSeq := schema.AlterSequenceChange{
		Sequence: &schema.Sequence{
			Name: "order_seq",
		},
	}
	_, err = my.GenerateSQL(alterSeq)
	require.Error(t, err)

	// Test dropping sequence (not directly supported in MySQL)
	dropSeq := schema.DropSequenceChange{
		SequenceName: "order_seq",
	}
	_, err = my.GenerateSQL(dropSeq)
	require.Error(t, err)
}

func TestCreateTable(t *testing.T) {
	my := New()

	// Test simple table creation
	table := &schema.Table{
		Name: "users",
		Columns: []*schema.Column{
			{Name: "id", Type: &schema.IntegerType{}, Nullable: false},
			{Name: "name", Type: &schema.VarcharType{Length: 100}, Nullable: false},
			{Name: "email", Type: &schema.VarcharType{Length: 255}, Nullable: false},
			{Name: "bio", Type: &schema.TextType{}, Nullable: true, Comment: "User biography"},
			{Name: "created_at", Type: &schema.TimestampType{}, Nullable: false, Default: "CURRENT_TIMESTAMP"},
		},
		PrimaryKey: &schema.PrimaryKey{
			Columns: []string{"id"},
		},
	}
	createTable := schema.CreateTableChange{
		TableDef: table,
	}
	sql, err := my.GenerateSQL(createTable)
	require.NoError(t, err)
	expected := "CREATE TABLE `users` (\n  `id` int NOT NULL,\n  `name` varchar(100) NOT NULL,\n  `email` varchar(255) NOT NULL,\n  `bio` text,\n  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,\n  PRIMARY KEY (`id`)\n) ENGINE=InnoDB;\nALTER TABLE `users` MODIFY COLUMN `bio` text COMMENT 'User biography';"
	require.Equal(t, testutil.FormatSQL(expected), testutil.FormatSQL(sql))
}

func TestDropTable(t *testing.T) {
	my := New()
	dropTable := schema.DropTableChange{
		TableName: "users",
	}
	sql, err := my.GenerateSQL(dropTable)
	require.NoError(t, err)
	require.Equal(t, "DROP TABLE `users`;", sql)
}

func TestAddColumn(t *testing.T) {
	my := New()
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
	sql, err := my.GenerateSQL(addColumn)
	require.NoError(t, err)
	require.Equal(t, "ALTER TABLE `users` ADD COLUMN `email` varchar(255) NOT NULL DEFAULT 'user@example.com' COMMENT 'User email address';", sql)
}

func TestDropColumn(t *testing.T) {
	my := New()
	dropColumn := schema.DropColumnChange{
		TableName:  "users",
		ColumnName: "email",
	}
	sql, err := my.GenerateSQL(dropColumn)
	require.NoError(t, err)
	require.Equal(t, "ALTER TABLE `users` DROP COLUMN `email`;", sql)
}

func TestAlterColumn(t *testing.T) {
	my := New()
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
	sql, err := my.GenerateSQL(alterColumn)
	require.NoError(t, err)
	require.Equal(t, "ALTER TABLE `users` MODIFY COLUMN `email` varchar(100) COMMENT 'Updated comment';", sql)
}

func TestAddPrimaryKey(t *testing.T) {
	my := New()

	// Test simple primary key
	addPK := schema.AddPrimaryKeyChange{
		TableName: "users",
		PrimaryKey: &schema.PrimaryKey{
			Columns: []string{"id"},
		},
	}
	sql, err := my.GenerateSQL(addPK)
	require.NoError(t, err)
	require.Equal(t, "ALTER TABLE `users` ADD PRIMARY KEY (`id`);", sql)

	// Test composite primary key
	addCompositePK := schema.AddPrimaryKeyChange{
		TableName: "order_items",
		PrimaryKey: &schema.PrimaryKey{
			Columns: []string{"order_id", "product_id"},
		},
	}
	sql, err = my.GenerateSQL(addCompositePK)
	require.NoError(t, err)
	require.Equal(t, "ALTER TABLE `order_items` ADD PRIMARY KEY (`order_id`, `product_id`);", sql)
}

func TestDropPrimaryKey(t *testing.T) {
	my := New()
	dropPK := schema.DropPrimaryKeyChange{
		TableName: "users",
	}
	sql, err := my.GenerateSQL(dropPK)
	require.NoError(t, err)
	require.Equal(t, "ALTER TABLE `users` DROP PRIMARY KEY;", sql)
}

func TestAddIndex(t *testing.T) {
	my := New()

	// Test simple index
	addIdx := schema.AddIndexChange{
		TableName: "users",
		Index: &schema.Index{
			Name:    "idx_users_email",
			Columns: []string{"email"},
			Unique:  false,
		},
	}
	sql, err := my.GenerateSQL(addIdx)
	require.NoError(t, err)
	require.Equal(t, "CREATE INDEX `idx_users_email` ON `users` (`email`);", sql)

	// Test unique index
	addUniqueIdx := schema.AddIndexChange{
		TableName: "users",
		Index: &schema.Index{
			Name:    "idx_users_email_unique",
			Columns: []string{"email"},
			Unique:  true,
		},
	}
	sql, err = my.GenerateSQL(addUniqueIdx)
	require.NoError(t, err)
	require.Equal(t, "CREATE UNIQUE INDEX `idx_users_email_unique` ON `users` (`email`);", sql)

	// Test multi-column index
	addMultiIdx := schema.AddIndexChange{
		TableName: "users",
		Index: &schema.Index{
			Name:    "idx_users_name_email",
			Columns: []string{"name", "email"},
		},
	}
	sql, err = my.GenerateSQL(addMultiIdx)
	require.NoError(t, err)
	require.Equal(t, "CREATE INDEX `idx_users_name_email` ON `users` (`name`, `email`);", sql)
}

func TestDropIndex(t *testing.T) {
	my := New()
	dropIdx := schema.DropIndexChange{
		TableName: "users",
		IndexName: "idx_users_email",
	}
	sql, err := my.GenerateSQL(dropIdx)
	require.NoError(t, err)
	require.Equal(t, "DROP INDEX `idx_users_email` ON `users`;", sql)
}

func TestAddForeignKey(t *testing.T) {
	my := New()

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
	sql, err := my.GenerateSQL(addFK)
	require.NoError(t, err)
	require.Equal(t, "ALTER TABLE `posts` ADD CONSTRAINT `fk_posts_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE;", sql)

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
	sql, err = my.GenerateSQL(addFKUpdate)
	require.NoError(t, err)
	require.Equal(t, "ALTER TABLE `posts` ADD CONSTRAINT `fk_posts_category` FOREIGN KEY (`category_id`) REFERENCES `categories` (`id`) ON UPDATE CASCADE;", sql)

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
	sql, err = my.GenerateSQL(addCompositeFK)
	require.NoError(t, err)
	require.Equal(t, "ALTER TABLE `order_items` ADD CONSTRAINT `fk_order_items` FOREIGN KEY (`order_id`, `product_id`) REFERENCES `products` (`order_id`, `id`);", sql)
}

func TestDropForeignKey(t *testing.T) {
	my := New()
	dropFK := schema.DropForeignKeyChange{
		TableName: "posts",
		FKName:    "fk_posts_user",
	}
	sql, err := my.GenerateSQL(dropFK)
	require.NoError(t, err)
	require.Equal(t, "ALTER TABLE `posts` DROP FOREIGN KEY `fk_posts_user`;", sql)
}

func TestCreateFunction(t *testing.T) {
	my := New()
	createFn := schema.CreateFunctionChange{
		Function: &schema.Function{
			Name: "add_numbers",
			Arguments: []schema.FunctionArg{
				{Name: "a", Type: "int"},
				{Name: "b", Type: "int"},
			},
			Returns:    "int",
			Volatility: "IMMUTABLE",
			Body:       "RETURN a + b;",
		},
	}
	sql, err := my.GenerateSQL(createFn)
	require.NoError(t, err)
	expected := "CREATE FUNCTION `add_numbers`(a int, b int)\nRETURNS int\nDETERMINISTIC\nBEGIN\nRETURN a + b;\nEND;"
	require.Equal(t, testutil.FormatSQL(expected), testutil.FormatSQL(sql))
}

func TestAlterFunction(t *testing.T) {
	my := New()
	alterFn := schema.AlterFunctionChange{
		Function: &schema.Function{
			Name: "add_numbers",
			Arguments: []schema.FunctionArg{
				{Name: "a", Type: "int"},
				{Name: "b", Type: "int"},
			},
			Returns:    "int",
			Volatility: "STABLE",
			Body:       "RETURN a + b + 1;",
		},
	}
	sql, err := my.GenerateSQL(alterFn)
	require.NoError(t, err)
	require.Contains(t, sql, "DROP FUNCTION IF EXISTS `add_numbers`")
	require.Contains(t, sql, "CREATE FUNCTION `add_numbers`")
}

func TestDropFunction(t *testing.T) {
	my := New()
	dropFn := schema.DropFunctionChange{
		FunctionName: "add_numbers",
	}
	sql, err := my.GenerateSQL(dropFn)
	require.NoError(t, err)
	require.Equal(t, "DROP FUNCTION IF EXISTS `add_numbers`;", sql)
}

func TestCreateView(t *testing.T) {
	my := New()

	// Test view with columns
	createView := schema.CreateViewChange{
		View: &schema.View{
			Name:       "active_users",
			Columns:    []string{"id", "name", "email"},
			Definition: "SELECT id, name, email FROM users WHERE active = 1",
		},
	}
	sql, err := my.GenerateSQL(createView)
	require.NoError(t, err)
	expected := "CREATE VIEW `active_users` (`id`, `name`, `email`) AS\nSELECT id, name, email FROM users WHERE active = 1;"
	require.Equal(t, testutil.FormatSQL(expected), testutil.FormatSQL(sql))

	// Test simple view
	simpleView := schema.CreateViewChange{
		View: &schema.View{
			Name:       "all_users",
			Definition: "SELECT * FROM users",
		},
	}
	sql, err = my.GenerateSQL(simpleView)
	require.NoError(t, err)
	require.Equal(t, "CREATE VIEW `all_users` AS\nSELECT * FROM users;", sql)
}

func TestAlterView(t *testing.T) {
	my := New()
	alterView := schema.AlterViewChange{
		View: &schema.View{
			Name:       "active_users",
			Columns:    []string{"id", "name", "email"},
			Definition: "SELECT id, name, email FROM users WHERE active = 1 AND verified = 1",
		},
	}
	sql, err := my.GenerateSQL(alterView)
	require.NoError(t, err)
	expected := "CREATE OR REPLACE VIEW `active_users` (`id`, `name`, `email`) AS\nSELECT id, name, email FROM users WHERE active = 1 AND verified = 1;"
	require.Equal(t, testutil.FormatSQL(expected), testutil.FormatSQL(sql))
}

func TestDropView(t *testing.T) {
	my := New()
	dropView := schema.DropViewChange{
		ViewName: "active_users",
	}
	sql, err := my.GenerateSQL(dropView)
	require.NoError(t, err)
	require.Equal(t, "DROP VIEW IF EXISTS `active_users`;", sql)
}

func TestCreateTrigger(t *testing.T) {
	my := New()

	// Test simple trigger
	createTrigger := schema.CreateTriggerChange{
		Trigger: &schema.Trigger{
			Name:     "update_timestamp",
			Table:    "users",
			Events:   []string{"UPDATE"},
			Timing:   "BEFORE",
			ForEach:  "ROW",
			Function: "SET NEW.updated_at = NOW();",
		},
	}
	sql, err := my.GenerateSQL(createTrigger)
	require.NoError(t, err)
	expected := "CREATE TRIGGER `update_timestamp`\nBEFORE UPDATE ON `users`\nFOR EACH ROW\nBEGIN\n  -- Call function: SET NEW.updated_at = NOW();\n  -- Replace with actual trigger body\nEND;"
	require.Equal(t, testutil.FormatSQL(expected), testutil.FormatSQL(sql))

	// Test trigger with multiple events
	multiEventTrigger := schema.CreateTriggerChange{
		Trigger: &schema.Trigger{
			Name:     "log_changes",
			Table:    "products",
			Events:   []string{"INSERT", "UPDATE"},
			Timing:   "AFTER",
			ForEach:  "ROW",
			Function: "INSERT INTO audit_log VALUES (NULL, NOW());",
		},
	}
	sql, err = my.GenerateSQL(multiEventTrigger)
	require.NoError(t, err)
	require.Contains(t, sql, "AFTER INSERT OR UPDATE")
}

func TestAlterTrigger(t *testing.T) {
	my := New()
	alterTrigger := schema.AlterTriggerChange{
		Trigger: &schema.Trigger{
			Name:     "update_timestamp",
			Table:    "users",
			Events:   []string{"UPDATE", "INSERT"},
			Timing:   "BEFORE",
			ForEach:  "ROW",
			Function: "SET NEW.updated_at = NOW();",
		},
	}
	sql, err := my.GenerateSQL(alterTrigger)
	require.NoError(t, err)
	require.Contains(t, sql, "DROP TRIGGER IF EXISTS `update_timestamp`")
	require.Contains(t, sql, "CREATE TRIGGER `update_timestamp`")
}

func TestDropTrigger(t *testing.T) {
	my := New()
	dropTrigger := schema.DropTriggerChange{
		TriggerName: "update_timestamp",
	}
	sql, err := my.GenerateSQL(dropTrigger)
	require.NoError(t, err)
	require.Equal(t, "DROP TRIGGER IF EXISTS `update_timestamp`;", sql)
}

func TestQuoteIdentifier(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"users", "`users`"},
		{"user table", "`user table`"},
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
		{"test", "'test'"},
		{"test's string", "'test''s string'"},
		{"O'Reilly", "'O''Reilly'"},
	}

	for _, test := range tests {
		require.Equal(t, test.expected, quoteLiteral(test.input))
	}
}
