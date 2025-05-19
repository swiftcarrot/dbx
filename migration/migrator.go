package migration

import (
	"database/sql"
	"fmt"

	"github.com/swiftcarrot/dbx/mysql"
	"github.com/swiftcarrot/dbx/postgresql"
	"github.com/swiftcarrot/dbx/schema"
	"github.com/swiftcarrot/dbx/sqlite"
)

// Migrator runs migrations
type Migrator struct {
	db             *sql.DB
	registry       *Registry
	versionTracker *VersionTracker
}

// NewMigrator creates a new migrator
func NewMigrator(db *sql.DB, registry *Registry) *Migrator {
	versionTracker := NewVersionTracker(db)

	return &Migrator{
		db:             db,
		registry:       registry,
		versionTracker: versionTracker,
	}
}

// Init initializes the migrator
func (m *Migrator) Init() error {
	return m.versionTracker.EnsureMigrationsTable()
}

// Migrate runs migrations up to the specified version
// If version is empty, all pending migrations are run
func (m *Migrator) Migrate(version string) error {
	// Initialize the migrations table if it doesn't exist
	if err := m.Init(); err != nil {
		return err
	}

	// Get all migrations
	migrations := m.registry.GetMigrations()
	if len(migrations) == 0 {
		return fmt.Errorf("no migrations found")
	}

	// Get applied migrations
	appliedMigrations, err := m.versionTracker.GetAppliedMigrations()
	if err != nil {
		return err
	}

	// Create a set of applied migrations for quick lookup
	appliedSet := make(map[string]bool)
	for _, m := range appliedMigrations {
		appliedSet[m.Version] = true
	}

	// Determine the database type for SQL generation
	dbType, err := m.versionTracker.DatabaseType()
	if err != nil {
		return err
	}

	// Create appropriate SQL generator based on database type
	var sqlGenerator schema.SQLGenerator
	switch dbType {
	case "mysql":
		sqlGenerator = mysql.New()
	case "postgresql":
		sqlGenerator = postgresql.New()
	case "sqlite":
		sqlGenerator = sqlite.New()
	default:
		return fmt.Errorf("unsupported database type: %s", dbType)
	}

	// Begin a transaction
	tx, err := m.db.Begin()
	if err != nil {
		return err
	}

	// Rollback on error
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// If we have a target version, migrate up to that version
	// Otherwise, run all pending migrations
	for _, migration := range migrations {
		// Skip if already applied
		if appliedSet[migration.Version] {
			continue
		}

		// Stop if we've reached the target version
		if version != "" && migration.Version > version {
			break
		}

		fmt.Printf("Migrating up: %s_%s\n", migration.Version, migration.Name)

		// Run the migration
		upSchema := migration.Up()

		// Get current schema
		var currentSchema *schema.Schema
		if version == "" && len(appliedMigrations) == 0 {
			// If this is the first migration, use an empty schema
			currentSchema = schema.NewSchema()
		} else {
			// Otherwise, inspect the current schema
			currentSchema, err = sqlGenerator.Inspect(m.db)
			if err != nil {
				return fmt.Errorf("failed to inspect schema: %w", err)
			}
		}

		// Generate changes
		changes := schema.Diff(currentSchema, upSchema)

		// Apply changes
		for _, change := range changes {
			sql, err := sqlGenerator.GenerateSQL(change)
			if err != nil {
				return fmt.Errorf("failed to generate SQL: %w", err)
			}

			_, err = tx.Exec(sql)
			if err != nil {
				return fmt.Errorf("failed to execute SQL: %w", err)
			}
		}

		// Record the migration
		err = m.versionTracker.RecordMigration(migration.Version, migration.Name)
		if err != nil {
			return fmt.Errorf("failed to record migration: %w", err)
		}
	}

	// Commit the transaction
	return tx.Commit()
}

// Rollback rolls back the last applied migration
// If steps is specified, that many migrations are rolled back
func (m *Migrator) Rollback(steps int) error {
	// Ensure migrations table exists
	if err := m.Init(); err != nil {
		return err
	}

	// Get applied migrations
	appliedMigrations, err := m.versionTracker.GetAppliedMigrations()
	if err != nil {
		return err
	}

	if len(appliedMigrations) == 0 {
		return fmt.Errorf("no migrations to roll back")
	}

	// Determine how many migrations to roll back
	if steps <= 0 {
		steps = 1 // Default to rolling back just the last migration
	}

	if steps > len(appliedMigrations) {
		steps = len(appliedMigrations)
	}

	// Get migrations to roll back
	migrationsToRollback := appliedMigrations[len(appliedMigrations)-steps:]

	// Determine the database type for SQL generation
	dbType, err := m.versionTracker.DatabaseType()
	if err != nil {
		return err
	}

	// Create appropriate SQL generator based on database type
	var sqlGenerator schema.SQLGenerator
	switch dbType {
	case "mysql":
		sqlGenerator = mysql.New()
	case "postgresql":
		sqlGenerator = postgresql.New()
	case "sqlite":
		sqlGenerator = sqlite.New()
	default:
		return fmt.Errorf("unsupported database type: %s", dbType)
	}

	// Begin a transaction
	tx, err := m.db.Begin()
	if err != nil {
		return err
	}

	// Rollback on error
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Roll back migrations in reverse order (newest first)
	for i := len(migrationsToRollback) - 1; i >= 0; i-- {
		record := migrationsToRollback[i]

		// Find the migration
		migration := m.registry.FindMigrationByVersion(record.Version)
		if migration == nil {
			return fmt.Errorf("migration with version %s not found in registry", record.Version)
		}

		fmt.Printf("Rolling back: %s_%s\n", migration.Version, migration.Name)

		// Run the down migration
		downSchema := migration.Down()

		// Get current schema
		currentSchema, err := sqlGenerator.Inspect(m.db)
		if err != nil {
			return fmt.Errorf("failed to inspect schema: %w", err)
		}

		// Generate changes
		changes := schema.Diff(currentSchema, downSchema)

		// Apply changes
		for _, change := range changes {
			sql, err := sqlGenerator.GenerateSQL(change)
			if err != nil {
				return fmt.Errorf("failed to generate SQL: %w", err)
			}

			_, err = tx.Exec(sql)
			if err != nil {
				return fmt.Errorf("failed to execute SQL: %w", err)
			}
		}

		// Remove the migration record
		err = m.versionTracker.RemoveMigration(migration.Version)
		if err != nil {
			return fmt.Errorf("failed to remove migration record: %w", err)
		}
	}

	// Commit the transaction
	return tx.Commit()
}

// Status returns the status of all migrations
func (m *Migrator) Status() ([]struct {
	Version   string
	Name      string
	Status    string
	AppliedAt string
}, error) {
	// Ensure migrations table exists
	if err := m.Init(); err != nil {
		return nil, err
	}

	// Get all migrations
	migrations := m.registry.GetMigrations()

	// Get applied migrations
	appliedMigrations, err := m.versionTracker.GetAppliedMigrations()
	if err != nil {
		return nil, err
	}

	// Create a map of applied migrations
	appliedMap := make(map[string]MigrationRecord)
	for _, m := range appliedMigrations {
		appliedMap[m.Version] = m
	}

	// Create status list
	statusList := make([]struct {
		Version   string
		Name      string
		Status    string
		AppliedAt string
	}, len(migrations))

	for i, migration := range migrations {
		status := "Pending"
		appliedAt := ""

		if record, ok := appliedMap[migration.Version]; ok {
			status = "Applied"
			appliedAt = record.AppliedAt.Format("2006-01-02 15:04:05")
		}

		statusList[i] = struct {
			Version   string
			Name      string
			Status    string
			AppliedAt string
		}{
			Version:   migration.Version,
			Name:      migration.Name,
			Status:    status,
			AppliedAt: appliedAt,
		}
	}

	return statusList, nil
}
