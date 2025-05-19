# DBX Migration System

DBX Migrations is a Rails-like database migration system for Go applications using the `github.com/swiftcarrot/dbx` library. It provides a simple way to manage database schema changes across different database systems.

## Features

- Rails-like migration system with timestamped versions
- Support for MySQL, PostgreSQL, and SQLite databases
- Full schema comparison and diff calculation
- Automatic SQL generation for schema changes
- Migration rollback support
- Migration status reporting

## Installation

```bash
go get github.com/swiftcarrot/dbx
```

## Usage

### Command Line Interface

The DBX library includes a command-line tool for managing migrations:

```bash
# Generate a new migration
dbx generate create_users

# Run all pending migrations
dbx --database "postgres://user:pass@localhost/dbname" migrate

# Run migrations up to a specific version
dbx --database "postgres://user:pass@localhost/dbname" migrate 20250520000000

# Rollback the last migration
dbx --database "postgres://user:pass@localhost/dbname" rollback

# Rollback multiple migrations
dbx --database "postgres://user:pass@localhost/dbname" rollback 3

# Show migration status
dbx --database "postgres://user:pass@localhost/dbname" status

# Get help
dbx help
```

### In Your Go Code

You can also use the migration system programmatically in your Go code:

```go
package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
	"github.com/swiftcarrot/dbx/migration"
)

func main() {
	// Set the migrations directory
	migration.SetMigrationsDir("./migrations")

	// Connect to database
	db, err := sql.Open("postgres", "postgres://user:pass@localhost/dbname")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Run migrations
	err = migration.RunMigrations(db, "")
	if err != nil {
		log.Fatal(err)
	}

	// Get migration status
	status, err := migration.GetMigrationStatus(db)
	if err != nil {
		log.Fatal(err)
	}

	for _, s := range status {
		log.Printf("Migration %s (%s): %s", s.Version, s.Name, s.Status)
	}
}
```

## Writing Migrations

Migrations are Go files with `Up` and `Down` functions that define schema changes. When you generate a migration, a new file is created in your migrations directory with this structure:

```go
package migrations

import (
	"github.com/swiftcarrot/dbx/migration"
	"github.com/swiftcarrot/dbx/schema"
)

func init() {
	migration.Register("20250520000000", "create_users", up20250520000000, down20250520000000)
}

func up20250520000000() *schema.Schema {
	s := schema.NewSchema()

	s.CreateTable("users", func(t *schema.Table) {
		t.Column("id", &schema.IntegerType{}, schema.PrimaryKey)
		t.Column("username", &schema.VarcharType{Length: 255}, schema.NotNull)
		t.Column("email", &schema.VarcharType{Length: 255}, schema.NotNull)
		t.Column("created_at", &schema.TimestampType{})

		t.Index("idx_users_email", []string{"email"}, schema.Unique)
	})

	return s
}

func down20250520000000() *schema.Schema {
	s := schema.NewSchema()

	s.DropTable("users")

	return s
}
```

### Schema Definition API

The schema definition API is powerful and allows you to define complex schema changes:

```go
// Create a table
s.CreateTable("posts", func(t *schema.Table) {
	// Add columns
	t.Column("id", &schema.IntegerType{})
	t.Column("title", &schema.VarcharType{Length: 255})
	t.Column("body", &schema.TextType{})
	t.Column("user_id", &schema.IntegerType{})

	// Add a primary key
	t.SetPrimaryKey("pk_posts", []string{"id"})

	// Add a foreign key
	t.ForeignKey("fk_posts_user", []string{"user_id"}, "users", []string{"id"})

	// Add indexes
	t.Index("idx_posts_title", []string{"title"})
})

// Create a view
s.CreateView("active_users", "SELECT * FROM users WHERE active = TRUE")

// PostgreSQL-specific features
s.EnableExtension("uuid-ossp")
s.CreateSequence("user_id_seq")
```

## Migration Workflow

1. Generate a new migration: `dbx generate create_users`
2. Edit the migration file to define schema changes
3. Run the migration: `dbx --database "postgres://..." migrate`
4. If needed, roll back: `dbx --database "postgres://..." rollback`

## Database Support

- PostgreSQL: Full support for all PostgreSQL features
- MySQL: Support for tables, columns, indexes, foreign keys, views, triggers
- SQLite: Support for basic schema operations

## License

See the main DBX project license.
