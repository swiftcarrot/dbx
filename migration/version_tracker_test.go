package migration

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/swiftcarrot/dbx/internal/testutil"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := testutil.GetSQLiteTestConn()
	require.NoError(t, err)

	t.Cleanup(func() {
		_, err := db.Exec(`DROP TABLE IF EXISTS schema_migrations`)
		require.NoError(t, err)
	})

	return db
}

func TestVersionTracker(t *testing.T) {
	db := setupTestDB(t)
	tracker := NewVersionTracker(db)

	// Test creating the migrations table
	err := tracker.EnsureMigrationsTable()
	require.NoError(t, err)

	// Test getting applied migrations when empty
	migrations, err := tracker.GetAppliedMigrations()
	require.NoError(t, err)
	require.Empty(t, migrations)

	// Test recording a migration
	err = tracker.RecordMigration("20250520000000", "create_users")
	require.NoError(t, err)

	// Test has migration
	has, err := tracker.HasMigration("20250520000000")
	require.NoError(t, err)
	require.True(t, has)

	has, err = tracker.HasMigration("nonexistent")
	require.NoError(t, err)
	require.False(t, has)

	// Test getting current version
	version, err := tracker.GetCurrentVersion()
	require.NoError(t, err)
	require.Equal(t, "20250520000000", version)

	// Test getting applied migrations
	migrations, err = tracker.GetAppliedMigrations()
	require.NoError(t, err)
	require.Equal(t, 1, len(migrations))
	require.Equal(t, "20250520000000", migrations[0].Version)
	require.Equal(t, "create_users", migrations[0].Name)
	require.False(t, migrations[0].AppliedAt.IsZero())

	// Add another migration
	err = tracker.RecordMigration("20250520000001", "create_posts")
	require.NoError(t, err)

	// Test getting all migrations
	migrations, err = tracker.GetAppliedMigrations()
	require.NoError(t, err)
	require.Equal(t, 2, len(migrations))

	// Test getting current version (should be the latest)
	version, err = tracker.GetCurrentVersion()
	require.NoError(t, err)
	require.Equal(t, "20250520000001", version)

	// Test removing a migration
	err = tracker.RemoveMigration("20250520000001")
	require.NoError(t, err)

	// Verify it was removed
	has, err = tracker.HasMigration("20250520000001")
	require.NoError(t, err)
	require.False(t, has)

	// Test getting current version after removal
	version, err = tracker.GetCurrentVersion()
	require.NoError(t, err)
	require.Equal(t, "20250520000000", version)
}

func TestDatabaseType(t *testing.T) {
	db := setupTestDB(t)
	tracker := NewVersionTracker(db)

	dbType, err := tracker.DatabaseType()
	require.NoError(t, err)
	require.Equal(t, "sqlite", dbType)
}
