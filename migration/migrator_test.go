package migration

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/swiftcarrot/dbx/schema"
)

func TestMigrator(t *testing.T) {
	db := setupTestDB(t)

	// Create a registry with test migrations
	registry := NewRegistry(t.TempDir())

	// Create test migrations
	createUsersMigration := NewMigration("20250520000000", "create_users", func() *schema.Schema {
		s := schema.NewSchema()
		s.CreateTable("users", func(t *schema.Table) {
			t.Column("id", &schema.IntegerType{})
			t.Column("name", &schema.VarcharType{Length: 255})
			t.SetPrimaryKey("pk_users", []string{"id"})
		})
		return s
	}, func() *schema.Schema {
		s := schema.NewSchema()
		s.DropTable("users")
		return s
	})

	createPostsMigration := NewMigration("20250520000001", "create_posts", func() *schema.Schema {
		s := schema.NewSchema()
		s.CreateTable("posts", func(t *schema.Table) {
			t.Column("id", &schema.IntegerType{})
			t.Column("title", &schema.VarcharType{Length: 255})
			t.Column("user_id", &schema.IntegerType{})
			t.SetPrimaryKey("pk_posts", []string{"id"})
			t.ForeignKey("fk_posts_user", []string{"user_id"}, "users", []string{"id"})
		})
		return s
	}, func() *schema.Schema {
		s := schema.NewSchema()
		s.DropTable("posts")
		return s
	})

	// Add migrations to registry
	registry.AddMigration(createUsersMigration)
	registry.AddMigration(createPostsMigration)

	// Create migrator
	migrator := NewMigrator(db, registry)

	// Initialize migrations table
	err := migrator.Init()
	require.NoError(t, err)

	// Run migrations
	err = migrator.Migrate("")
	require.NoError(t, err)

	// Check that migrations were applied
	versionTracker := NewVersionTracker(db)

	migrations, err := versionTracker.GetAppliedMigrations()
	require.NoError(t, err)
	require.Equal(t, 2, len(migrations))

	// Verify tables exist
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='users'").Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count)

	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='posts'").Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count)

	// Test rolling back one migration
	err = migrator.Rollback(1)
	require.NoError(t, err)

	// Verify posts table was dropped
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='posts'").Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 0, count)

	// Verify users table still exists
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='users'").Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count)

	// Verify migration record was removed
	has, err := versionTracker.HasMigration("20250520000001")
	require.NoError(t, err)
	require.False(t, has)

	// Test partial migration (up to specific version)
	err = migrator.Migrate("20250520000001")
	require.NoError(t, err)

	// Verify posts table exists again
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='posts'").Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count)

	// Test status
	status, err := migrator.Status()
	require.NoError(t, err)
	require.Equal(t, 2, len(status))
	require.Equal(t, "Applied", status[0].Status)
	require.Equal(t, "Applied", status[1].Status)

	// Test rolling back all migrations
	err = migrator.Rollback(2)
	require.NoError(t, err)

	// Verify both tables were dropped
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name IN ('users', 'posts')").Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 0, count)
}
