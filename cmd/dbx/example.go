package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/swiftcarrot/dbx/migration"
	"github.com/swiftcarrot/dbx/schema"
)

// This example demonstrates how to use the migration system
func Example() {
	// Set the migrations directory
	migrationsDir := filepath.Join(os.TempDir(), "dbx_migrations")
	migration.SetMigrationsDir(migrationsDir)

	// Create a new migration
	migrationPath, err := migration.CreateMigration("create_users")
	if err != nil {
		fmt.Printf("Error creating migration: %s\n", err)
		return
	}
	fmt.Printf("Created migration: %s\n", migrationPath)

	// Register a migration programmatically
	migration.Register("20250520000000", "create_posts", upCreatePosts, downCreatePosts)
}

// Migration functions for create_posts
func upCreatePosts() *schema.Schema {
	s := schema.NewSchema()

	s.CreateTable("posts", func(t *schema.Table) {
		t.Column("id", &schema.IntegerType{})
		t.Column("title", &schema.VarcharType{Length: 255})
		t.Column("body", &schema.TextType{})
		t.Column("user_id", &schema.IntegerType{})
		t.Column("created_at", &schema.TimestampType{})
		t.Column("updated_at", &schema.TimestampType{})

		t.SetPrimaryKey("pk_posts", []string{"id"})
		t.ForeignKey("fk_posts_user", []string{"user_id"}, "users", []string{"id"})

		t.Index("idx_posts_created_at", []string{"created_at"})
	})

	return s
}

func downCreatePosts() *schema.Schema {
	s := schema.NewSchema()

	s.DropTable("posts")

	return s
}
