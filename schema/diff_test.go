package schema

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDiff(t *testing.T) {
	tests := []struct {
		name     string
		source   *Schema
		target   *Schema
		expected []Change
	}{
		{
			name:     "Empty schemas",
			source:   NewSchema(),
			target:   NewSchema(),
			expected: []Change{},
		},
		{
			name:   "Create table",
			source: NewSchema(),
			target: func() *Schema {
				s := NewSchema()
				s.CreateTable("users", func(t *Table) {
					t.Column("id", "integer")
					t.Column("name", "varchar")
				})
				return s
			}(),
			expected: []Change{
				&CreateTableChange{
					TableDef: &Table{
						Name: "users",
						Columns: []*Column{
							{Name: "id", Type: "integer"},
							{Name: "name", Type: "varchar"},
						},
						Indexes: []*Index{},
					},
				},
			},
		},
		{
			name: "Drop table",
			source: func() *Schema {
				s := NewSchema()
				s.CreateTable("users", func(t *Table) {
					t.Column("id", "integer")
				})
				return s
			}(),
			target: NewSchema(),
			expected: []Change{
				&DropTableChange{
					TableName: "users",
				},
			},
		},
		{
			name: "Add column",
			source: func() *Schema {
				s := NewSchema()
				s.CreateTable("users", func(t *Table) {
					t.Column("id", "integer")
				})
				return s
			}(),
			target: func() *Schema {
				s := NewSchema()
				s.CreateTable("users", func(t *Table) {
					t.Column("id", "integer")
					t.Column("name", "varchar")
				})
				return s
			}(),
			expected: []Change{
				&AddColumnChange{
					TableName: "users",
					Column: &Column{
						Name: "name",
						Type: "varchar",
					},
				},
			},
		},
		{
			name: "Drop column",
			source: func() *Schema {
				s := NewSchema()
				s.CreateTable("users", func(t *Table) {
					t.Column("id", "integer")
					t.Column("name", "varchar")
				})
				return s
			}(),
			target: func() *Schema {
				s := NewSchema()
				s.CreateTable("users", func(t *Table) {
					t.Column("id", "integer")
				})
				return s
			}(),
			expected: []Change{
				&DropColumnChange{
					TableName:  "users",
					ColumnName: "name",
				},
			},
		},
		{
			name: "Alter column",
			source: func() *Schema {
				s := NewSchema()
				s.CreateTable("users", func(t *Table) {
					t.Column("id", "integer")
					t.Column("name", "varchar")
				})
				return s
			}(),
			target: func() *Schema {
				s := NewSchema()
				s.CreateTable("users", func(t *Table) {
					t.Column("id", "integer")
					t.Column("name", "text")
				})
				return s
			}(),
			expected: []Change{
				&AlterColumnChange{
					TableName: "users",
					Column: &Column{
						Name: "name",
						Type: "text",
					},
				},
			},
		},
		{
			name: "Add primary key",
			source: func() *Schema {
				s := NewSchema()
				s.CreateTable("users", func(t *Table) {
					t.Column("id", "integer")
				})
				return s
			}(),
			target: func() *Schema {
				s := NewSchema()
				s.CreateTable("users", func(t *Table) {
					t.Column("id", "integer")
					t.SetPrimaryKey("pk_users", []string{"id"})
				})
				return s
			}(),
			expected: []Change{
				&AddPrimaryKeyChange{
					TableName: "users",
					PrimaryKey: &PrimaryKey{
						Name:    "pk_users",
						Columns: []string{"id"},
					},
				},
			},
		},
		{
			name: "Drop primary key",
			source: func() *Schema {
				s := NewSchema()
				s.CreateTable("users", func(t *Table) {
					t.Column("id", "integer")
					t.SetPrimaryKey("pk_users", []string{"id"})
				})
				return s
			}(),
			target: func() *Schema {
				s := NewSchema()
				s.CreateTable("users", func(t *Table) {
					t.Column("id", "integer")
				})
				return s
			}(),
			expected: []Change{
				&DropPrimaryKeyChange{
					TableName: "users",
					PKName:    "pk_users",
				},
			},
		},
		{
			name: "Add index",
			source: func() *Schema {
				s := NewSchema()
				s.CreateTable("users", func(t *Table) {
					t.Column("id", "integer")
					t.Column("name", "varchar")
				})
				return s
			}(),
			target: func() *Schema {
				s := NewSchema()
				s.CreateTable("users", func(t *Table) {
					t.Column("id", "integer")
					t.Column("name", "varchar")
					t.Index("idx_name", []string{"name"})
				})
				return s
			}(),
			expected: []Change{
				&AddIndexChange{
					TableName: "users",
					Index: &Index{
						Name:    "idx_name",
						Columns: []string{"name"},
					},
				},
			},
		},
		{
			name: "Drop index",
			source: func() *Schema {
				s := NewSchema()
				s.CreateTable("users", func(t *Table) {
					t.Column("id", "integer")
					t.Column("name", "varchar")
					t.Index("idx_name", []string{"name"})
				})
				return s
			}(),
			target: func() *Schema {
				s := NewSchema()
				s.CreateTable("users", func(t *Table) {
					t.Column("id", "integer")
					t.Column("name", "varchar")
				})
				return s
			}(),
			expected: []Change{
				&DropIndexChange{
					TableName: "users",
					IndexName: "idx_name",
				},
			},
		},
		{
			name: "Add foreign key",
			source: func() *Schema {
				s := NewSchema()
				s.CreateTable("users", func(t *Table) {
					t.Column("id", "integer")
				})
				s.CreateTable("posts", func(t *Table) {
					t.Column("id", "integer")
					t.Column("user_id", "integer")
				})
				return s
			}(),
			target: func() *Schema {
				s := NewSchema()
				s.CreateTable("users", func(t *Table) {
					t.Column("id", "integer")
				})
				s.CreateTable("posts", func(t *Table) {
					t.Column("id", "integer")
					t.Column("user_id", "integer")
					t.ForeignKey("fk_user", []string{"user_id"}, "users", []string{"id"})
				})
				return s
			}(),
			expected: []Change{
				&AddForeignKeyChange{
					TableName: "posts",
					ForeignKey: &ForeignKey{
						Name:       "fk_user",
						Columns:    []string{"user_id"},
						RefTable:   "users",
						RefColumns: []string{"id"},
					},
				},
			},
		},
		{
			name: "Drop foreign key",
			source: func() *Schema {
				s := NewSchema()
				s.CreateTable("users", func(t *Table) {
					t.Column("id", "integer")
				})
				s.CreateTable("posts", func(t *Table) {
					t.Column("id", "integer")
					t.Column("user_id", "integer")
					t.ForeignKey("fk_user", []string{"user_id"}, "users", []string{"id"})
				})
				return s
			}(),
			target: func() *Schema {
				s := NewSchema()
				s.CreateTable("users", func(t *Table) {
					t.Column("id", "integer")
				})
				s.CreateTable("posts", func(t *Table) {
					t.Column("id", "integer")
					t.Column("user_id", "integer")
				})
				return s
			}(),
			expected: []Change{
				&DropForeignKeyChange{
					TableName: "posts",
					FKName:    "fk_user",
				},
			},
		},
		{
			name: "Change primary key",
			source: func() *Schema {
				s := NewSchema()
				s.CreateTable("users", func(t *Table) {
					t.Column("id", "integer")
					t.Column("uuid", "varchar")
					t.SetPrimaryKey("pk_users_id", []string{"id"})
				})
				return s
			}(),
			target: func() *Schema {
				s := NewSchema()
				s.CreateTable("users", func(t *Table) {
					t.Column("id", "integer")
					t.Column("uuid", "varchar")
					t.SetPrimaryKey("pk_users_uuid", []string{"uuid"})
				})
				return s
			}(),
			expected: []Change{
				&DropPrimaryKeyChange{
					TableName: "users",
					PKName:    "pk_users_id",
				},
				&AddPrimaryKeyChange{
					TableName: "users",
					PrimaryKey: &PrimaryKey{
						Name:    "pk_users_uuid",
						Columns: []string{"uuid"},
					},
				},
			},
		},
		{
			name: "Multiple changes",
			source: func() *Schema {
				s := NewSchema()
				s.CreateTable("users", func(t *Table) {
					t.Column("id", "integer")
					t.Column("old_name", "varchar")
				})
				s.CreateTable("old_table", func(t *Table) {
					t.Column("id", "integer")
				})
				return s
			}(),
			target: func() *Schema {
				s := NewSchema()
				s.CreateTable("users", func(t *Table) {
					t.Column("id", "integer")
					t.Column("name", "varchar")
					t.SetPrimaryKey("pk_users", []string{"id"})
				})
				s.CreateTable("new_table", func(t *Table) {
					t.Column("id", "integer")
				})
				return s
			}(),
			expected: []Change{
				&DropTableChange{
					TableName: "old_table",
				},
				&DropColumnChange{
					TableName:  "users",
					ColumnName: "old_name",
				},
				&AddColumnChange{
					TableName: "users",
					Column: &Column{
						Name: "name",
						Type: "varchar",
					},
				},
				&AddPrimaryKeyChange{
					TableName: "users",
					PrimaryKey: &PrimaryKey{
						Name:    "pk_users",
						Columns: []string{"id"},
					},
				},
				&CreateTableChange{
					TableDef: &Table{
						Name:    "new_table",
						Columns: []*Column{{Name: "id", Type: "integer"}},
						Indexes: []*Index{},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		changes := Diff(tt.source, tt.target)
		require.Equal(t, tt.expected, changes, tt.name)
	}
}
