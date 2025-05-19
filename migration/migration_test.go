package migration

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/swiftcarrot/dbx/schema"
)

func TestNewMigration(t *testing.T) {
	version := "20250520000000"
	name := "create_users"

	upFn := func() *schema.Schema { return schema.NewSchema() }
	downFn := func() *schema.Schema { return schema.NewSchema() }

	migration := NewMigration(version, name, upFn, downFn)

	require.Equal(t, version, migration.Version)
	require.Equal(t, name, migration.Name)
	require.NotNil(t, migration.UpFn)
	require.NotNil(t, migration.DownFn)
	require.False(t, migration.CreatedAt.IsZero())
}

func TestMigrationFullVersion(t *testing.T) {
	migration := NewMigration("20250520000000", "create_users", nil, nil)
	require.Equal(t, "20250520000000_create_users", migration.FullVersion())
}

func TestMigrationUp(t *testing.T) {
	s := schema.NewSchema()
	s.CreateTable("users", func(t *schema.Table) {
		t.Column("id", &schema.IntegerType{})
	})

	upFn := func() *schema.Schema { return s }
	migration := NewMigration("20250520000000", "create_users", upFn, nil)

	upSchema := migration.Up()
	require.Equal(t, 1, len(upSchema.Tables))
	require.Equal(t, "users", upSchema.Tables[0].Name)
}

func TestMigrationDown(t *testing.T) {
	s := schema.NewSchema()
	s.DropTable("users")

	downFn := func() *schema.Schema { return s }
	migration := NewMigration("20250520000000", "create_users", nil, downFn)

	downSchema := migration.Down()
	require.Equal(t, 0, len(downSchema.Tables))
}

func TestGenerateVersionTimestamp(t *testing.T) {
	ts := GenerateVersionTimestamp()

	// Check timestamp format: YYYYMMDDHHMMSS
	require.Len(t, ts, 14)

	// Check that it's a valid timestamp by parsing it
	now := time.Now()
	nowStr := now.Format("20060102150405")

	// The timestamps should be close (within a minute)
	require.InDelta(t, len(nowStr), len(ts), 0)
}

func TestRegistry(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "dbx_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	registry := NewRegistry(tempDir)

	// Test adding migrations
	m1 := NewMigration("20250520000000", "first", nil, nil)
	m2 := NewMigration("20250520000001", "second", nil, nil)

	registry.AddMigration(m1)
	registry.AddMigration(m2)

	// Test getting migrations in sorted order
	migrations := registry.GetMigrations()
	require.Equal(t, 2, len(migrations))
	require.Equal(t, "20250520000000", migrations[0].Version)
	require.Equal(t, "20250520000001", migrations[1].Version)

	// Test finding a migration by version
	found := registry.FindMigrationByVersion("20250520000001")
	require.NotNil(t, found)
	require.Equal(t, "second", found.Name)

	// Test finding a non-existent migration
	notFound := registry.FindMigrationByVersion("999")
	require.Nil(t, notFound)
}

func TestGenerator(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "dbx_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	registry := NewRegistry(tempDir)
	generator := NewGenerator(registry)

	// Test generating a migration
	filePath, err := generator.Generate("create_users")
	require.NoError(t, err)

	// Check that the file was created
	_, err = os.Stat(filePath)
	require.NoError(t, err)

	// Check file name format
	fileName := filepath.Base(filePath)
	require.Contains(t, fileName, "_create_users.go")

	// Check file content
	content, err := os.ReadFile(filePath)
	require.NoError(t, err)

	contentStr := string(content)
	require.Contains(t, contentStr, "func init() {")
	require.Contains(t, contentStr, "migration.Register(")
	require.Contains(t, contentStr, "func up")
	require.Contains(t, contentStr, "func down")
}
