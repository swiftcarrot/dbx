# dbx

> [!WARNING]
> This project is in beta. The API is subject to changes and may break.

[![Go Reference](https://pkg.go.dev/badge/github.com/swiftcarrot/dbx.svg)](https://pkg.go.dev/github.com/swiftcarrot/dbx)
[![Go Report Card](https://goreportcard.com/badge/github.com/swiftcarrot/dbx)](https://goreportcard.com/report/github.com/swiftcarrot/dbx)
[![test](https://github.com/swiftcarrot/dbx/actions/workflows/test.yml/badge.svg?branch=main)](https://github.com/swiftcarrot/dbx/actions/workflows/test.yml)
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/swiftcarrot/dbx)

dbx is a database schema migration library for Go that lets you manage database schemas using Go code instead of SQL.

## Features

- **Database inspection**: Introspect existing database schemas
- **Schema comparison**: Compare schemas and generate migration statements
- **Built on `database/sql`**: Works with standard Go database drivers
- **Automatic SQL generation**: Automatically generate SQL statements for schema changes

## Usage Examples

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
