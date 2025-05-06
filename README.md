# dbx

[![Go Reference](https://pkg.go.dev/badge/github.com/swiftcarrot/dbx.svg)](https://pkg.go.dev/github.com/swiftcarrot/dbx)
[![Go Report Card](https://goreportcard.com/badge/github.com/swiftcarrot/dbx)](https://goreportcard.com/report/github.com/swiftcarrot/dbx)
[![CI Status](https://github.com/swiftcarrot/dbx/workflows/CI/badge.svg)](https://github.com/swiftcarrot/dbx/actions)

DBX is a powerful database schema migration library for Go that enables programmatic management of database schemas using Go code instead of raw SQL.

## Features

- **Schema-based migrations**: Define your database schema in Go code rather than raw SQL
- **Automatic SQL generation**: Automatically generate SQL statements for schema changes
- **Database inspection**: Introspect existing database schemas
- **Schema comparison**: Compare schemas and generate migration statements
- **Built on `database/sql`**: Works with standard Go database drivers
- **Multiple dialect support**: PostgreSQL and MySQL supported, with more coming soon

## Usage Examples

### Database Inspection

Introspect an existing database schema:

```go
import (
	_ "github.com/lib/pq"
	"github.com/swiftcarrot/dbx/postgresql"
	"github.com/swiftcarrot/dbx/schema"
)

db, err := sql.Open("postgres", "postgres://")
pg := postgresql.New()
source, err := pg.Inspect(db)
```

### Schema Definition and Comparison

Define a target schema and compare with current schema:

```go
target := schema.NewSchema()
target.CreateTable("user", func(t *schema.Table) {
	t.Column("name", "text", schema.NotNull)
	t.Index("")
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

For other dialect support, feel free to [create an issue](https://github.com/swiftcarrot/dbx/issues/new).

## License

This project is licensed under the Apache License - see the LICENSE file for details.
