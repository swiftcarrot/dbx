package migration

import (
	"database/sql"
	"sync"

	"github.com/swiftcarrot/dbx/schema"
)

// Global registry for migrations
var (
	defaultRegistry *Registry
	defaultOnce     sync.Once
)

// SetMigrationsDir sets the directory where migrations are stored
func SetMigrationsDir(dir string) {
	defaultOnce.Do(func() {
		defaultRegistry = NewRegistry(dir)
	})
}

// Register registers a new migration in the default registry
func Register(version, name string, upFn, downFn func() *schema.Schema) {
	if defaultRegistry == nil {
		panic("migration directory not set; call SetMigrationsDir first")
	}

	migration := NewMigration(version, name, upFn, downFn)
	defaultRegistry.AddMigration(migration)
}

// CreateMigration generates a new migration file
func CreateMigration(name string) (string, error) {
	if defaultRegistry == nil {
		panic("migration directory not set; call SetMigrationsDir first")
	}

	generator := NewGenerator(defaultRegistry)
	return generator.Generate(name)
}

// RunMigrations runs pending migrations
func RunMigrations(db *sql.DB, targetVersion string) error {
	if defaultRegistry == nil {
		panic("migration directory not set; call SetMigrationsDir first")
	}

	if err := defaultRegistry.LoadMigrations(); err != nil {
		return err
	}

	migrator := NewMigrator(db, defaultRegistry)
	return migrator.Migrate(targetVersion)
}

// RollbackMigration rolls back the last migration or specified number of migrations
func RollbackMigration(db *sql.DB, steps int) error {
	if defaultRegistry == nil {
		panic("migration directory not set; call SetMigrationsDir first")
	}

	if err := defaultRegistry.LoadMigrations(); err != nil {
		return err
	}

	migrator := NewMigrator(db, defaultRegistry)
	return migrator.Rollback(steps)
}

// GetMigrationStatus returns the status of all migrations
func GetMigrationStatus(db *sql.DB) ([]struct {
	Version   string
	Name      string
	Status    string
	AppliedAt string
}, error) {
	if defaultRegistry == nil {
		panic("migration directory not set; call SetMigrationsDir first")
	}

	if err := defaultRegistry.LoadMigrations(); err != nil {
		return nil, err
	}

	migrator := NewMigrator(db, defaultRegistry)
	return migrator.Status()
}

// GetCurrentVersion returns the current database version
func GetCurrentVersion(db *sql.DB) (string, error) {
	versionTracker := NewVersionTracker(db)
	if err := versionTracker.EnsureMigrationsTable(); err != nil {
		return "", err
	}

	return versionTracker.GetCurrentVersion()
}
