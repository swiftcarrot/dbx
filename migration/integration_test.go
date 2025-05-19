package migration

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/swiftcarrot/dbx/internal/testutil"
	"github.com/swiftcarrot/dbx/schema"
)

func TestIntegration(t *testing.T) {
	// Create a temporary directory for migrations
	migrationsDir, err := os.MkdirTemp("", "dbx_integration_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(migrationsDir)

	// Set up registry
	registry := NewRegistry(migrationsDir)

	// Generate migrations
	generator := NewGenerator(registry)

	createUsersPath, err := generator.Generate("create_users")
	require.NoError(t, err)

	createPostsPath, err := generator.Generate("create_posts")
	require.NoError(t, err)

	// Verify migration files were created
	_, err = os.Stat(createUsersPath)
	require.NoError(t, err)

	_, err = os.Stat(createPostsPath)
	require.NoError(t, err)

	// Extract the version from the filename
	usersVersion := filepath.Base(createUsersPath)
	usersVersion = usersVersion[:14] // Get timestamp part

	// Update the users migration content with the correct version
	usersMigrationContent := `package migrations

import (
	"github.com/swiftcarrot/dbx/migration"
	"github.com/swiftcarrot/dbx/schema"
)

func init() {
	migration.Register("` + usersVersion + `", "create_users", upCreateUsers, downCreateUsers)
}

func upCreateUsers() *schema.Schema {
	s := schema.NewSchema()
	s.CreateTable("users", func(t *schema.Table) {
		t.Column("id", &schema.IntegerType{})
		t.Column("name", &schema.VarcharType{Length: 255})
		t.Column("email", &schema.VarcharType{Length: 255})
		t.Column("created_at", &schema.TimestampType{})
		t.SetPrimaryKey("pk_users", []string{"id"})
	})
	return s
}

func downCreateUsers() *schema.Schema {
	s := schema.NewSchema()
	s.DropTable("users")
	return s
}
`

	err = os.WriteFile(createUsersPath, []byte(usersMigrationContent), 0644)
	require.NoError(t, err)

	postsVersion := filepath.Base(createPostsPath)
	postsVersion = postsVersion[:14] // Get timestamp part

	postsMigrationContent := `package migrations

import (
	"github.com/swiftcarrot/dbx/migration"
	"github.com/swiftcarrot/dbx/schema"
)

func init() {
	migration.Register("` + postsVersion + `", "create_posts", upCreatePosts, downCreatePosts)
}

func upCreatePosts() *schema.Schema {
	s := schema.NewSchema()
	s.CreateTable("posts", func(t *schema.Table) {
		t.Column("id", &schema.IntegerType{})
		t.Column("title", &schema.VarcharType{Length: 255})
		t.Column("content", &schema.TextType{})
		t.Column("user_id", &schema.IntegerType{})
		t.Column("created_at", &schema.TimestampType{})
		t.SetPrimaryKey("pk_posts", []string{"id"})
		t.ForeignKey("fk_posts_user", []string{"user_id"}, "users", []string{"id"})
	})
	return s
}

func downCreatePosts() *schema.Schema {
	s := schema.NewSchema()
	s.DropTable("posts")
	return s
}
`

	err = os.WriteFile(createPostsPath, []byte(postsMigrationContent), 0644)
	require.NoError(t, err)

	// Create a test database
	db, err := testutil.GetSQLiteTestConn()
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err := db.Exec(`
			DROP TABLE IF EXISTS posts;
			DROP TABLE IF EXISTS users;
			DROP TABLE IF EXISTS schema_migrations;
		`)
		require.NoError(t, err)
	})

	// Register migrations programmatically since we can't rely on init() during testing
	versionTracker := NewVersionTracker(db)
	err = versionTracker.EnsureMigrationsTable()
	require.NoError(t, err)

	// We need to manually create migration objects since loading Go files dynamically is complex
	createUsersFn := func() *schema.Schema {
		s := schema.NewSchema()
		s.CreateTable("users", func(t *schema.Table) {
			t.Column("id", &schema.IntegerType{})
			t.Column("name", &schema.VarcharType{Length: 255})
			t.Column("email", &schema.VarcharType{Length: 255})
			t.Column("created_at", &schema.TimestampType{})
			t.SetPrimaryKey("pk_users", []string{"id"})
		})
		return s
	}

	dropUsersFn := func() *schema.Schema {
		s := schema.NewSchema()
		s.DropTable("users")
		return s
	}

	createPostsFn := func() *schema.Schema {
		s := schema.NewSchema()
		s.CreateTable("posts", func(t *schema.Table) {
			t.Column("id", &schema.IntegerType{})
			t.Column("title", &schema.VarcharType{Length: 255})
			t.Column("content", &schema.TextType{})
			t.Column("user_id", &schema.IntegerType{})
			t.Column("created_at", &schema.TimestampType{})
			t.SetPrimaryKey("pk_posts", []string{"id"})
			t.ForeignKey("fk_posts_user", []string{"user_id"}, "users", []string{"id"})
		})
		return s
	}

	dropPostsFn := func() *schema.Schema {
		s := schema.NewSchema()
		s.DropTable("posts")
		return s
	}

	usersMigration := NewMigration(usersVersion, "create_users", createUsersFn, dropUsersFn)
	postsMigration := NewMigration(postsVersion, "create_posts", createPostsFn, dropPostsFn)

	registry.AddMigration(usersMigration)
	registry.AddMigration(postsMigration)

	// Create migrator and run migrations
	migrator := NewMigrator(db, registry)

	// Run first migration only
	err = migrator.Migrate(usersVersion)
	require.NoError(t, err)

	// Verify first migration was applied
	var tableCount int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='users'").Scan(&tableCount)
	require.NoError(t, err)
	require.Equal(t, 1, tableCount)

	// Run second migration
	err = migrator.Migrate("")
	require.NoError(t, err)

	// Verify second migration was applied
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='posts'").Scan(&tableCount)
	require.NoError(t, err)
	require.Equal(t, 1, tableCount)

	// Check that the foreign key exists
	var fkCount int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM sqlite_master
		WHERE type='table' AND name='posts'
		AND sql LIKE '%FOREIGN KEY%REFERENCES%users%'
	`).Scan(&fkCount)
	require.NoError(t, err)
	require.Equal(t, 1, fkCount)

	// Check migration records
	migrations, err := versionTracker.GetAppliedMigrations()
	require.NoError(t, err)
	require.Equal(t, 2, len(migrations))

	// Test rollback
	err = migrator.Rollback(1)
	require.NoError(t, err)

	// Verify posts table was dropped
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='posts'").Scan(&tableCount)
	require.NoError(t, err)
	require.Equal(t, 0, tableCount)

	// Verify users table still exists
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='users'").Scan(&tableCount)
	require.NoError(t, err)
	require.Equal(t, 1, tableCount)
}

	// Generate migrations
	generator := NewGenerator(registry)

	createUsersPath, err := generator.Generate("create_users")
	require.NoError(t, err)

	createPostsPath, err := generator.Generate("create_posts")
	require.NoError(t, err)

	// Verify migration files were created
	_, err = os.Stat(createUsersPath)
	require.NoError(t, err)

	_, err = os.Stat(createPostsPath)
	require.NoError(t, err)

	// Update migration file content with actual schema changes
	usersMigrationContent := `package migrations

import (
	"github.com/swiftcarrot/dbx/migration"
	"github.com/swiftcarrot/dbx/schema"
)

func init() {
	version := filepath.Base(createUsersPath)
	version = version[:14] // Extract timestamp part

	migration.Register(version, "create_users", upCreateUsers, downCreateUsers)
}

func upCreateUsers() *schema.Schema {
	s := schema.NewSchema()
	s.CreateTable("users", func(t *schema.Table) {
		t.Column("id", &schema.IntegerType{})
		t.Column("name", &schema.VarcharType{Length: 255})
		t.Column("email", &schema.VarcharType{Length: 255})
		t.Column("created_at", &schema.TimestampType{})
		t.SetPrimaryKey("pk_users", []string{"id"})
	})
	return s
}

func downCreateUsers() *schema.Schema {
	s := schema.NewSchema()
	s.DropTable("users")
	return s
}
`
	// Extract the version from the filename
	usersVersion := filepath.Base(createUsersPath)
	usersVersion = usersVersion[:14] // Get timestamp part

	// Update the users migration content with the correct version
	usersMigrationContent = `package migrations

import (
	"github.com/swiftcarrot/dbx/migration"
	"github.com/swiftcarrot/dbx/schema"
)

func init() {
	migration.Register("` + usersVersion + `", "create_users", upCreateUsers, downCreateUsers)
}

func upCreateUsers() *schema.Schema {
	s := schema.NewSchema()
	s.CreateTable("users", func(t *schema.Table) {
		t.Column("id", &schema.IntegerType{})
		t.Column("name", &schema.VarcharType{Length: 255})
		t.Column("email", &schema.VarcharType{Length: 255})
		t.Column("created_at", &schema.TimestampType{})
		t.SetPrimaryKey("pk_users", []string{"id"})
	})
	return s
}

func downCreateUsers() *schema.Schema {
	s := schema.NewSchema()
	s.DropTable("users")
	return s
}
`

	err = os.WriteFile(createUsersPath, []byte(usersMigrationContent), 0644)
	require.NoError(t, err)

	postsVersion := filepath.Base(createPostsPath)
	postsVersion = postsVersion[:14] // Get timestamp part

	postsMigrationContent := `package migrations

import (
	"github.com/swiftcarrot/dbx/migration"
	"github.com/swiftcarrot/dbx/schema"
)

func init() {
	migration.Register("` + postsVersion + `", "create_posts", upCreatePosts, downCreatePosts)
}

func upCreatePosts() *schema.Schema {
	s := schema.NewSchema()
	s.CreateTable("posts", func(t *schema.Table) {
		t.Column("id", &schema.IntegerType{})
		t.Column("title", &schema.VarcharType{Length: 255})
		t.Column("content", &schema.TextType{})
		t.Column("user_id", &schema.IntegerType{})
		t.Column("created_at", &schema.TimestampType{})
		t.SetPrimaryKey("pk_posts", []string{"id"})
		t.ForeignKey("fk_posts_user", []string{"user_id"}, "users", []string{"id"})
	})
	return s
}

func downCreatePosts() *schema.Schema {
	s := schema.NewSchema()
	s.DropTable("posts")
	return s
}
`

	err = os.WriteFile(createPostsPath, []byte(postsMigrationContent), 0644)
	require.NoError(t, err)

	// Create a test database
	db, err := testutil.GetSQLiteTestConn()
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err := db.Exec(`
			DROP TABLE IF EXISTS posts;
			DROP TABLE IF EXISTS users;
			DROP TABLE IF EXISTS schema_migrations;
		`)
		require.NoError(t, err)
	})

	// Register migrations programmatically since we can't rely on init() during testing
	versionTracker := NewVersionTracker(db)
	err = versionTracker.EnsureMigrationsTable()
	require.NoError(t, err)

	// We need to manually create migration objects since loading Go files dynamically is complex
	createUsersFn := func() *schema.Schema {
		s := schema.NewSchema()
		s.CreateTable("users", func(t *schema.Table) {
			t.Column("id", &schema.IntegerType{})
			t.Column("name", &schema.VarcharType{Length: 255})
			t.Column("email", &schema.VarcharType{Length: 255})
			t.Column("created_at", &schema.TimestampType{})
			t.SetPrimaryKey("pk_users", []string{"id"})
		})
		return s
	}

	dropUsersFn := func() *schema.Schema {
		s := schema.NewSchema()
		s.DropTable("users")
		return s
	}

	createPostsFn := func() *schema.Schema {
		s := schema.NewSchema()
		s.CreateTable("posts", func(t *schema.Table) {
			t.Column("id", &schema.IntegerType{})
			t.Column("title", &schema.VarcharType{Length: 255})
			t.Column("content", &schema.TextType{})
			t.Column("user_id", &schema.IntegerType{})
			t.Column("created_at", &schema.TimestampType{})
			t.SetPrimaryKey("pk_posts", []string{"id"})
			t.ForeignKey("fk_posts_user", []string{"user_id"}, "users", []string{"id"})
		})
		return s
	}

	dropPostsFn := func() *schema.Schema {
		s := schema.NewSchema()
		s.DropTable("posts")
		return s
	}

	usersMigration := NewMigration(usersVersion, "create_users", createUsersFn, dropUsersFn)
	postsMigration := NewMigration(postsVersion, "create_posts", createPostsFn, dropPostsFn)

	registry.AddMigration(usersMigration)
	registry.AddMigration(postsMigration)

	// Create migrator and run migrations
	migrator := NewMigrator(db, registry)

	// Run first migration only
	err = migrator.Migrate(usersVersion)
	require.NoError(t, err)

	// Verify first migration was applied
	var tableCount int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='users'").Scan(&tableCount)
	require.NoError(t, err)
	require.Equal(t, 1, tableCount)

	// Run second migration
	err = migrator.Migrate("")
	require.NoError(t, err)

	// Verify second migration was applied
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='posts'").Scan(&tableCount)
	require.NoError(t, err)
	require.Equal(t, 1, tableCount)

	// Check that the foreign key exists
	var fkCount int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM sqlite_master
		WHERE type='table' AND name='posts'
		AND sql LIKE '%FOREIGN KEY%REFERENCES%users%'
	`).Scan(&fkCount)
	require.NoError(t, err)
	require.Equal(t, 1, fkCount)

	// Check migration records
	migrations, err := versionTracker.GetAppliedMigrations()
	require.NoError(t, err)
	require.Equal(t, 2, len(migrations))

	// Test rollback
	err = migrator.Rollback(1)
	require.NoError(t, err)

	// Verify posts table was dropped
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='posts'").Scan(&tableCount)
	require.NoError(t, err)
	require.Equal(t, 0, tableCount)

	// Verify users table still exists
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='users'").Scan(&tableCount)
	require.NoError(t, err)
	require.Equal(t, 1, tableCount)
}
