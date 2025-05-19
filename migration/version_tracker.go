package migration

import (
	"database/sql"
	"fmt"
	"time"
)

// VersionTracker manages the migration versions in the database
type VersionTracker struct {
	db *sql.DB
}

// MigrationRecord represents a record in the schema_migrations table
type MigrationRecord struct {
	Version   string
	Name      string
	AppliedAt time.Time
}

// NewVersionTracker creates a new version tracker
func NewVersionTracker(db *sql.DB) *VersionTracker {
	return &VersionTracker{
		db: db,
	}
}

// EnsureMigrationsTable ensures that the schema_migrations table exists
func (v *VersionTracker) EnsureMigrationsTable() error {
	// Create the schema_migrations table if it doesn't exist
	// This table tracks which migrations have been applied
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version VARCHAR(255) PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		applied_at TIMESTAMP NOT NULL
	)`

	_, err := v.db.Exec(createTableSQL)
	return err
}

// GetAppliedMigrations returns all applied migrations
func (v *VersionTracker) GetAppliedMigrations() ([]MigrationRecord, error) {
	rows, err := v.db.Query("SELECT version, name, applied_at FROM schema_migrations ORDER BY version")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var migrations []MigrationRecord
	for rows.Next() {
		var m MigrationRecord
		if err := rows.Scan(&m.Version, &m.Name, &m.AppliedAt); err != nil {
			return nil, err
		}
		migrations = append(migrations, m)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return migrations, nil
}

// RecordMigration records a migration as applied
func (v *VersionTracker) RecordMigration(version, name string) error {
	_, err := v.db.Exec(
		"INSERT INTO schema_migrations (version, name, applied_at) VALUES (?, ?, ?)",
		version, name, time.Now(),
	)
	return err
}

// RemoveMigration removes a migration record
func (v *VersionTracker) RemoveMigration(version string) error {
	_, err := v.db.Exec("DELETE FROM schema_migrations WHERE version = ?", version)
	return err
}

// GetCurrentVersion returns the highest applied migration version
func (v *VersionTracker) GetCurrentVersion() (string, error) {
	var version string
	err := v.db.QueryRow("SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1").Scan(&version)
	if err == sql.ErrNoRows {
		return "", nil // No migrations applied yet
	}
	if err != nil {
		return "", err
	}
	return version, nil
}

// HasMigration checks if a migration has been applied
func (v *VersionTracker) HasMigration(version string) (bool, error) {
	var count int
	err := v.db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", version).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// DatabaseType determines the type of database (mysql, postgresql, sqlite)
func (v *VersionTracker) DatabaseType() (string, error) {
	// Try to determine database type from the driver name
	// This is a simple approach and might need refinement
	driverName := v.db.Driver().Name()

	switch {
	case driverName == "mysql" || driverName == "mysqld":
		return "mysql", nil
	case driverName == "postgres" || driverName == "postgresql":
		return "postgresql", nil
	case driverName == "sqlite" || driverName == "sqlite3":
		return "sqlite", nil
	default:
		return "", fmt.Errorf("unsupported database type: %s", driverName)
	}
}
