package testutil

import (
	"github.com/swiftcarrot/dbx/schema"
)

// CreateUsersTableSchema creates a schema with a users table for testing.
func CreateUsersTableSchema() *schema.Schema {
	s := schema.NewSchema()
	s.CreateTable("users", func(t *schema.Table) {
		t.Column("id", "serial", schema.NotNull)
		t.Column("username", "varchar(50)", schema.NotNull)
		t.Column("email", "varchar(100)", schema.NotNull)
		t.Column("created_at", "timestamp", schema.NotNull, schema.Default("CURRENT_TIMESTAMP"))
		t.SetPrimaryKey("users_pkey", []string{"id"})
		t.Index("users_email_idx", []string{"email"}, schema.Unique)
		t.Index("users_username_idx", []string{"username"})
	})
	return s
}

// CreateUsersTableWithProfileSchema creates a schema with a users table that includes profile fields.
func CreateUsersTableWithProfileSchema() *schema.Schema {
	s := CreateUsersTableSchema()
	table := s.Tables[0] // users table
	table.Column("bio", "text", schema.Nullable)
	table.Column("avatar_url", "varchar(255)", schema.Nullable)
	return s
}

// CreateUsersAndPostsSchema creates a schema with users and posts tables.
func CreateUsersAndPostsSchema() *schema.Schema {
	s := CreateUsersTableSchema()
	s.CreateTable("posts", func(t *schema.Table) {
		t.Column("id", "serial", schema.NotNull)
		t.Column("user_id", "integer", schema.NotNull)
		t.Column("title", "varchar(200)", schema.NotNull)
		t.Column("content", "text", schema.Nullable)
		t.Column("published", "boolean", schema.NotNull, schema.Default("false"))
		t.Column("created_at", "timestamp", schema.NotNull, schema.Default("CURRENT_TIMESTAMP"))
		t.SetPrimaryKey("posts_pkey", []string{"id"})
		t.Index("posts_user_id_idx", []string{"user_id"})
		t.Index("posts_published_created_at_idx", []string{"published", "created_at"})
		t.ForeignKey("fk_posts_user", []string{"user_id"}, "users", []string{"id"}, schema.OnDelete("CASCADE"))
	})
	return s
}

// CreateFullSchema creates a schema with users, posts, and comments tables.
func CreateFullSchema() *schema.Schema {
	s := CreateUsersAndPostsSchema()
	s.CreateTable("comments", func(t *schema.Table) {
		t.Column("post_id", "integer", schema.NotNull)
		t.Column("user_id", "integer", schema.NotNull)
		t.Column("content", "text", schema.NotNull)
		t.Column("created_at", "timestamp", schema.NotNull, schema.Default("CURRENT_TIMESTAMP"))
		t.SetPrimaryKey("comments_pkey", []string{"post_id", "user_id"})
		t.ForeignKey("fk_comments_post", []string{"post_id"}, "posts", []string{"id"}, schema.OnDelete("CASCADE"))
		t.ForeignKey("fk_comments_user", []string{"user_id"}, "users", []string{"id"}, schema.OnDelete("CASCADE"))
	})
	return s
}

// CreateUsersTableWithModifiedColumnSchema creates a users schema with modified columns.
func CreateUsersTableWithModifiedColumnSchema() *schema.Schema {
	s := schema.NewSchema()
	s.CreateTable("users", func(t *schema.Table) {
		t.Column("id", "serial", schema.NotNull)
		t.Column("username", "varchar(50)", schema.Nullable) // Changed to nullable
		t.Column("email", "varchar(255)", schema.NotNull)    // Changed length
		t.Column("created_at", "timestamp", schema.NotNull, schema.Default("CURRENT_TIMESTAMP"))
		t.SetPrimaryKey("users_pkey", []string{"id"})
		t.Index("users_email_idx", []string{"email"}, schema.Unique)
		t.Index("users_username_idx", []string{"username"})
	})
	return s
}
