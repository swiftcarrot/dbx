# dbx

> [!WARNING]
> This project is in beta. The API is subject to changes and may break.

[![Go Reference](https://pkg.go.dev/badge/github.com/swiftcarrot/dbx.svg)](https://pkg.go.dev/github.com/swiftcarrot/dbx)
[![Go Report Card](https://goreportcard.com/badge/github.com/swiftcarrot/dbx)](https://goreportcard.com/report/github.com/swiftcarrot/dbx)
[![test](https://github.com/swiftcarrot/dbx/actions/workflows/test.yml/badge.svg?branch=main)](https://github.com/swiftcarrot/dbx/actions/workflows/test.yml)
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/swiftcarrot/dbx)

dbx is a database schema migration library for Go that lets you manage database schemas using Go code instead of SQL.

## Features

- **Rails-like migrations**: Define and organize schema changes with timestamp-based migrations
- **Database inspection**: Introspect existing database schemas
- **Schema comparison**: Compare schemas and generate migration statements
- **Built on `database/sql`**: Works with standard Go database drivers
- **Automatic SQL generation**: Automatically generate SQL statements for schema changes
- **CLI tool**: Command-line interface for creating and running migrations

## Usage Examples

### Rails-like Migrations

Create and run migrations using the CLI:

```bash
# Generate a new migration
dbx generate create_users

# Run migrations
dbx --database "postgres://postgres:postgres@localhost:5432/dbx_test?sslmode=disable" migrate
```

Or programmatically in your code:

```go
import (
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/swiftcarrot/dbx/migration"
)

// Set the migrations directory
migration.SetMigrationsDir("./migrations")

// Run migrations
db, _ := sql.Open("postgres", "postgres://postgres:postgres@localhost:5432/dbname")
migration.RunMigrations(db, "")
```

See the [migration documentation](./migration/README.md) for more details.

### Database Inspection

Introspect an existing database schema:

```go
import (
	_ "github.com/lib/pq"
	"github.com/swiftcarrot/dbx/postgresql"
	"github.com/swiftcarrot/dbx/schema"
)

db, err := sql.Open("postgres", "postgres://postgres:postgres@localhost:5432/dbx_test?sslmode=disable")
pg := postgresql.New()
source, err := pg.Inspect(db)
```

### Schema Definition and Comparison

Define a target schema and compare with current schema:

```go
target := schema.NewSchema()
target.CreateTable("user", func(t *schema.Table) {
	t.Column("name", &schema.TextType{}, schema.NotNull)
	t.Index("users_name_idx", []string{"name"})
})

changes, err := schema.Diff(source, target)
```

### Applying Schema Changes

Generate and execute SQL from schema changes:

```go
for _, change := range changes {
	sql := pg.GenerateSQL(change)
	_, err := db.Exec(sql)
}
```

## Data Types

You can define columns using convenient predefined helpers, or use the generic `Column` method with any supported type. All types accept optional column options (e.g., `NotNull`, `Default(...)`).

| Go Method  | SQL Type     | Example Usage              |
| ---------- | ------------ | -------------------------- |
| `String`   | VARCHAR(255) | `t.String("name")`         |
| `Text`     | TEXT         | `t.Text("bio")`            |
| `Integer`  | INTEGER      | `t.Integer("age")`         |
| `BigInt`   | BIGINT       | `t.BigInt("credit")`       |
| `Float`    | FLOAT        | `t.Float("weight")`        |
| `Decimal`  | DECIMAL      | `t.Decimal("balance")`     |
| `DateTime` | TIMESTAMP    | `t.DateTime("created_at")` |
| `Time`     | TIME         | `t.Time("time")`           |
| `Date`     | DATE         | `t.Date("birthday")`       |
| `Binary`   | BLOB/BINARY  | `t.Binary("bin")`          |
| `Boolean`  | BOOLEAN      | `t.Boolean("verified")`    |

You can also use the generic method:

```go
t.Column("column_name", &schema.TextType{}, schema.NotNull)
```

Example:

```go
sch.CreateTable("users", func (t *schema.Table) {
	t.String("name")
	t.Text("bio")
	t.Integer("age")
	t.BigInt("credit")
	t.Float("weight")
	t.Decimal("balance")
	t.DateTime("created_at")
	t.Time("time")
	t.Date("birthday")
	t.Binary("bin")
	t.Boolean("verified")
})
```

## Supported Dialects

### PostgreSQL

```go
import (
	_ "github.com/lib/pq"
	"github.com/swiftcarrot/dbx/postgresql"
)

pg := postgresql.New()
```

### MySQL

```go
import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/swiftcarrot/dbx/mysql"
)

my := mysql.New()
```

### SQLite

```go
import (
	_ "github.com/mattn/go-sqlite3"
	"github.com/swiftcarrot/dbx/sqlite"
)

s := sqlite.New()
```

For other dialect support, feel free to [create an issue](https://github.com/swiftcarrot/dbx/issues/new).

## Documentation and Support

The official dbx documentation is currently in development. In the meantime, you can:

- Read and ask questions on the [DeepWiki page](https://deepwiki.com/swiftcarrot/dbx)
- [Open an issue](https://github.com/swiftcarrot/dbx/issues) for bug reports or feature requests
- Join our [Discord community](https://discord.gg/t9y7gQBYem) for real-time discussions and support

## License

This project is licensed under the Apache License - see the LICENSE file for details.
