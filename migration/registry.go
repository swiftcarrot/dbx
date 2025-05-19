package migration

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Registry manages migration files and versions
type Registry struct {
	migrations    []*Migration
	migrationsDir string
}

// NewRegistry creates a new migration registry
func NewRegistry(migrationsDir string) *Registry {
	return &Registry{
		migrations:    []*Migration{},
		migrationsDir: migrationsDir,
	}
}

// AddMigration adds a migration to the registry
func (r *Registry) AddMigration(migration *Migration) {
	r.migrations = append(r.migrations, migration)
}

// FindMigrationByVersion finds a migration by its version
func (r *Registry) FindMigrationByVersion(version string) *Migration {
	for _, m := range r.migrations {
		if m.Version == version {
			return m
		}
	}
	return nil
}

// GetMigrations returns all migrations sorted by version
func (r *Registry) GetMigrations() []*Migration {
	// Sort migrations by version (which should be a timestamp)
	sort.Slice(r.migrations, func(i, j int) bool {
		return r.migrations[i].Version < r.migrations[j].Version
	})

	return r.migrations
}

// LoadMigrations loads all migrations from the migrations directory
func (r *Registry) LoadMigrations() error {
	// Check if migrations directory exists
	if _, err := os.Stat(r.migrationsDir); os.IsNotExist(err) {
		// Create migrations directory if it doesn't exist
		if err := os.MkdirAll(r.migrationsDir, 0755); err != nil {
			return fmt.Errorf("failed to create migrations directory: %w", err)
		}
		return nil
	}

	// Walk through the migrations directory
	err := filepath.WalkDir(r.migrationsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Only process .go files
		if !strings.HasSuffix(d.Name(), ".go") {
			return nil
		}

		// Skip test files
		if strings.HasSuffix(d.Name(), "_test.go") {
			return nil
		}

		// Skip the migrations.go file itself
		if d.Name() == "migrations.go" {
			return nil
		}

		// TODO: Parse migration file and register the migration
		// This would be done by the migration generator

		return nil
	})

	return err
}

// GenerateVersionTimestamp generates a timestamp-based version
// in the format used by Rails migrations (YYYYMMDDHHMMSS)
func GenerateVersionTimestamp() string {
	now := time.Now()
	return fmt.Sprintf("%d%02d%02d%02d%02d%02d",
		now.Year(), now.Month(), now.Day(),
		now.Hour(), now.Minute(), now.Second())
}

// GetMigrationsDir returns the migrations directory
func (r *Registry) GetMigrationsDir() string {
	return r.migrationsDir
}
