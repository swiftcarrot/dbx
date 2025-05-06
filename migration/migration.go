package migration

import (
	"time"

	"github.com/swiftcarrot/dbx/schema"
)

// Migration represents a database schema migration similar to Rails migrations.
// It includes version information, name, and methods to define schema changes.
type Migration struct {
	// Version represents the migration version (typically a timestamp)
	Version string

	// Name is a descriptive name for the migration
	Name string

	// CreatedAt represents when the migration was created
	CreatedAt time.Time

	// UpFn defines the schema changes for migrating up
	UpFn func() *schema.Schema

	// DownFn defines the schema changes for rolling back (migrating down)
	DownFn func() *schema.Schema
}

// NewMigration creates a new migration with the given version and name
func NewMigration(version, name string, upFn, downFn func() *schema.Schema) *Migration {
	return &Migration{
		Version:   version,
		Name:      name,
		CreatedAt: time.Now(),
		UpFn:      upFn,
		DownFn:    downFn,
	}
}

// Up returns the schema for migrating up
func (m *Migration) Up() *schema.Schema {
	if m.UpFn != nil {
		return m.UpFn()
	}
	return schema.NewSchema()
}

// Down returns the schema for migrating down (rolling back)
func (m *Migration) Down() *schema.Schema {
	if m.DownFn != nil {
		return m.DownFn()
	}
	return schema.NewSchema()
}

// FullVersion returns the combined version and name (e.g., "20250504120000_create_users")
func (m *Migration) FullVersion() string {
	return m.Version + "_" + m.Name
}
