package sqlite

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/swiftcarrot/dbx/internal/testutil"
	"github.com/swiftcarrot/dbx/schema"
)

func TestInspectIndexes(t *testing.T) {
	db, err := testutil.GetSQLiteTestConn()
	require.NoError(t, err)

	_, err = db.Exec(`
		CREATE TABLE test_index_users (
			id INTEGER PRIMARY KEY,
			email TEXT NOT NULL,
			username TEXT NOT NULL,
			first_name TEXT,
			last_name TEXT
		);

		CREATE UNIQUE INDEX idx_users_email ON test_index_users (email);
		CREATE INDEX idx_users_username ON test_index_users (username);
		CREATE INDEX idx_users_name ON test_index_users (first_name, last_name);
	`)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err := db.Exec(`DROP TABLE IF EXISTS test_index_users`)
		require.NoError(t, err)
	})

	s := New()
	table := &schema.Table{
		Name: "test_index_users",
	}
	err = s.InspectIndexes(db, table)
	require.NoError(t, err)
	require.Equal(t, []*schema.Index{
		{
			Name:    "idx_users_email",
			Columns: []string{"email"},
			Unique:  true,
		},
		{
			Name:    "idx_users_username",
			Columns: []string{"username"},
			Unique:  false,
		},
		{
			Name:    "idx_users_name",
			Columns: []string{"first_name", "last_name"},
			Unique:  false,
		},
	}, table.Indexes)
}
