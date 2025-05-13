package sqlite

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/swiftcarrot/dbx/internal/testutil"
)

func TestInspect(t *testing.T) {
	db, err := testutil.GetSQLiteTestConn()
	require.NoError(t, err)

	_, err = db.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT NOT NULL,
			bio TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		CREATE UNIQUE INDEX idx_users_email ON users (email);

		CREATE TABLE posts (
			id INTEGER PRIMARY KEY,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			user_id INTEGER NOT NULL,
			published BOOLEAN NOT NULL DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		);

		CREATE INDEX idx_posts_user_id ON posts (user_id);

		CREATE VIEW active_users AS
		SELECT id, name, email FROM users
		WHERE id IN (SELECT user_id FROM posts WHERE published = 1);

		CREATE TRIGGER update_posts_timestamp
		AFTER UPDATE OF title, content ON posts
		FOR EACH ROW
		BEGIN
			UPDATE posts SET created_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
		END;
	`)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err := db.Exec(`
			DROP TRIGGER IF EXISTS update_posts_timestamp;
			DROP VIEW IF EXISTS active_users;
			DROP TABLE IF EXISTS posts;
			DROP TABLE IF EXISTS users;
		`)
		require.NoError(t, err)
	})

	s := New()
	schema, err := s.Inspect(db)
	require.NoError(t, err)

	// Verify tables
	require.Equal(t, []string{"posts", "users"}, []string{schema.Tables[0].Name, schema.Tables[1].Name})

	// Verify users table
	usersTable := schema.Tables[1]
	require.Equal(t, "users", usersTable.Name)
	require.Equal(t, 5, len(usersTable.Columns))
	require.Equal(t, "id", usersTable.PrimaryKey.Columns[0])
	require.Equal(t, 1, len(usersTable.Indexes))
	require.Equal(t, "idx_users_email", usersTable.Indexes[0].Name)
	require.True(t, usersTable.Indexes[0].Unique)

	// Verify posts table
	postsTable := schema.Tables[0]
	require.Equal(t, "posts", postsTable.Name)
	require.Equal(t, 6, len(postsTable.Columns))
	require.Equal(t, "id", postsTable.PrimaryKey.Columns[0])
	require.Equal(t, 1, len(postsTable.Indexes))
	require.Equal(t, 1, len(postsTable.ForeignKeys))
	require.Equal(t, "users", postsTable.ForeignKeys[0].RefTable)
	require.Equal(t, "CASCADE", postsTable.ForeignKeys[0].OnDelete)

	// Verify views
	require.Equal(t, 1, len(schema.Views))
	require.Equal(t, "active_users", schema.Views[0].Name)

	// Verify triggers
	require.Equal(t, 1, len(schema.Triggers))
	require.Equal(t, "update_posts_timestamp", schema.Triggers[0].Name)
	require.Equal(t, "posts", schema.Triggers[0].Table)
}
