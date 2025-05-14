package mysql

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/swiftcarrot/dbx/internal/testutil"
	"github.com/swiftcarrot/dbx/schema"
)

func TestInspectForeignKeys(t *testing.T) {
	db, err := testutil.GetMySQLTestConn()
	require.NoError(t, err)

	_, err = db.Exec(`
		CREATE TABLE users (
			id INT PRIMARY KEY,
			name VARCHAR(255)
		);

		CREATE TABLE posts (
			id INT PRIMARY KEY,
			title VARCHAR(255),
			user_id INT,
			CONSTRAINT fk_posts_users FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		);

		CREATE TABLE comments (
			id INT PRIMARY KEY,
			content TEXT,
			user_id INT,
			post_id INT,
			CONSTRAINT fk_comments_users FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
			CONSTRAINT fk_comments_posts FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE
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

	my := New()

	postsTable := &schema.Table{
		Name: "posts",
	}
	err = my.InspectForeignKeys(db, postsTable)
	require.NoError(t, err)
	require.Equal(t, []*schema.ForeignKey{
		{
			Name:       "fk_posts_users",
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
	err = my.InspectForeignKeys(db, commentsTable)
	require.NoError(t, err)
	require.Equal(t, []*schema.ForeignKey{
		{
			Name:       "fk_comments_posts",
			Columns:    []string{"post_id"},
			RefTable:   "posts",
			RefColumns: []string{"id"},
			OnDelete:   "CASCADE",
			OnUpdate:   "NO ACTION",
		},
		{
			Name:       "fk_comments_users",
			Columns:    []string{"user_id"},
			RefTable:   "users",
			RefColumns: []string{"id"},
			OnDelete:   "SET NULL",
			OnUpdate:   "NO ACTION",
		},
	}, commentsTable.ForeignKeys)
}
