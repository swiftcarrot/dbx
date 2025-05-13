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
		CREATE TABLE test_users (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL
		);

		CREATE TABLE test_posts (
			id INTEGER PRIMARY KEY,
			user_id INTEGER NOT NULL,
			title TEXT NOT NULL,
			FOREIGN KEY (user_id) REFERENCES test_users (id) ON DELETE CASCADE
		);

		CREATE TABLE test_comments (
			id INTEGER PRIMARY KEY,
			post_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			content TEXT NOT NULL,
			FOREIGN KEY (post_id) REFERENCES test_posts (id) ON DELETE CASCADE ON UPDATE RESTRICT,
			FOREIGN KEY (user_id) REFERENCES test_users (id) ON DELETE CASCADE
		);
	`)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err := db.Exec(`
			DROP TABLE IF EXISTS test_comments;
			DROP TABLE IF EXISTS test_posts;
			DROP TABLE IF EXISTS test_users;
		`)
		require.NoError(t, err)
	})

	s := New()

	// Test table with no foreign keys
	tableNoFKs := &schema.Table{
		Name: "test_users",
	}
	err = s.InspectForeignKeys(db, tableNoFKs)
	require.NoError(t, err)
	require.Equal(t, []*schema.ForeignKey{}, tableNoFKs.ForeignKeys)

	// Test table with one foreign key
	tableSingleFK := &schema.Table{
		Name: "test_posts",
	}
	err = s.InspectForeignKeys(db, tableSingleFK)
	require.NoError(t, err)
	require.Equal(t, []*schema.ForeignKey{
		{
			Name:       "fk_test_posts_test_users",
			Columns:    []string{"user_id"},
			RefTable:   "test_users",
			RefColumns: []string{"id"},
			OnDelete:   "CASCADE",
			OnUpdate:   "NO ACTION",
		},
	}, tableSingleFK.ForeignKeys)

	// Test table with multiple foreign keys
	tableMultiFK := &schema.Table{
		Name: "test_comments",
	}
	err = s.InspectForeignKeys(db, tableMultiFK)
	require.NoError(t, err)
	require.Equal(t, []*schema.ForeignKey{
		{
			Name:       "fk_test_comments_test_posts",
			Columns:    []string{"post_id"},
			RefTable:   "test_posts",
			RefColumns: []string{"id"},
			OnDelete:   "CASCADE",
			OnUpdate:   "RESTRICT",
		},
		{
			Name:       "fk_test_comments_test_users",
			Columns:    []string{"user_id"},
			RefTable:   "test_users",
			RefColumns: []string{"id"},
			OnDelete:   "CASCADE",
			OnUpdate:   "NO ACTION",
		},
	}, tableMultiFK.ForeignKeys)
}
