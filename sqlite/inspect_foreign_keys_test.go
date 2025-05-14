package sqlite

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/swiftcarrot/dbx/internal/testutil"
	"github.com/swiftcarrot/dbx/schema"
)

func TestInspectForeignKeys(t *testing.T) {
	db, err := testutil.GetSQLiteTestConn()
	require.NoError(t, err)

	_, err = db.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL
		);

		CREATE TABLE posts (
			id INTEGER PRIMARY KEY,
			title TEXT NOT NULL,
			user_id INTEGER NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		);

		CREATE TABLE comments (
			id INTEGER PRIMARY KEY,
			content TEXT NOT NULL,
			user_id INTEGER,
			post_id INTEGER NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
			FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE
		);
	`)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err := db.Exec(`
			DROP TABLE IF EXISTS comments;
			DROP TABLE IF EXISTS posts;
			DROP TABLE IF EXISTS users;
		`)
		require.NoError(t, err)
	})

	s := New()
	postsTable := &schema.Table{
		Name: "posts",
	}
	err = s.InspectForeignKeys(db, postsTable)
	require.NoError(t, err)
	require.Equal(t, []*schema.ForeignKey{
		{
			Columns:    []string{"user_id"},
			RefTable:   "users",
			RefColumns: []string{"id"},
			OnDelete:   "CASCADE",
			OnUpdate:   "NO ACTION",
		},
	}, postsTable.ForeignKeys)

	commentsTable := &schema.Table{
		Name: "comments",
	}
	err = s.InspectForeignKeys(db, commentsTable)
	require.NoError(t, err)
	require.Equal(t, []*schema.ForeignKey{
		{
			Columns:    []string{"post_id"},
			RefTable:   "posts",
			RefColumns: []string{"id"},
			OnDelete:   "CASCADE",
			OnUpdate:   "NO ACTION",
		},
		{
			Columns:    []string{"user_id"},
			RefTable:   "users",
			RefColumns: []string{"id"},
			OnDelete:   "SET NULL",
			OnUpdate:   "NO ACTION",
		},
	}, commentsTable.ForeignKeys)
}
