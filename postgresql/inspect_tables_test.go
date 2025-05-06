package postgresql

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/swiftcarrot/dbx/internal/testutil"
)

func TestInspectTables(t *testing.T) {
	db, err := testutil.GetPGTestConn()
	require.NoError(t, err)

	_, err = db.Exec(`
		CREATE TABLE test_table_1 (id serial PRIMARY KEY);
		CREATE TABLE test_table_2 (id serial PRIMARY KEY);
	`)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err := db.Exec(`
			DROP TABLE IF EXISTS test_table_1;
			DROP TABLE IF EXISTS test_table_2;
			DROP TABLE IF EXISTS other_schema.test_table_3;
			DROP SCHEMA IF EXISTS other_schema;
		`)
		require.NoError(t, err)
	})

	pg := New()
	tables, err := pg.InspectTables(db)
	require.NoError(t, err)
	require.Equal(t, []string{"test_table_1", "test_table_2"}, tables)
}
