package sqlite

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/swiftcarrot/dbx/internal/testutil"
	"github.com/swiftcarrot/dbx/schema"
)

func TestInspectViews(t *testing.T) {
	db, err := testutil.GetSQLiteTestConn()
	require.NoError(t, err)

	_, err = db.Exec(`
		CREATE TABLE test_products (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			price REAL NOT NULL
		);

		CREATE VIEW view_test_simple AS SELECT id, name, price FROM test_products WHERE price > 0;
	`)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err := db.Exec(`
			DROP VIEW IF EXISTS view_test_simple;
			DROP TABLE IF EXISTS test_products;
		`)
		require.NoError(t, err)
	})

	s := New()
	schm := schema.NewSchema()
	err = s.InspectViews(db, schm)
	require.NoError(t, err)

	require.Equal(t, []*schema.View{
		{
			Name:       "view_test_simple",
			Definition: "SELECT id, name, price FROM test_products WHERE price > 0",
			Columns:    []string{"id", "name", "price"},
		},
	}, schm.Views)
}
