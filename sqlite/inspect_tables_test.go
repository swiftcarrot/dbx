package sqlite

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/swiftcarrot/dbx/internal/testutil"
)

func TestInspectTables(t *testing.T) {
	db, err := testutil.GetSQLiteTestConn()
	require.NoError(t, err)

	_, err = db.Exec(`
		CREATE TABLE test_table_1 (
			id INTEGER PRIMARY KEY
		);

		CREATE TABLE test_table_2 (
			id INTEGER PRIMARY KEY
		);
	`)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err := db.Exec(`
			DROP TABLE IF EXISTS test_table_1;
			DROP TABLE IF EXISTS test_table_2;
		`)
		require.NoError(t, err)
	})

	s := New()
	tables, err := s.InspectTables(db)
	require.NoError(t, err)
	require.Equal(t, []string{"test_table_1", "test_table_2"}, tables)
}
