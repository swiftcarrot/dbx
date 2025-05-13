package sqlite

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/swiftcarrot/dbx/internal/testutil"
	"github.com/swiftcarrot/dbx/schema"
)

func TestInspectPrimaryKey(t *testing.T) {
	db, err := testutil.GetSQLiteTestConn()
	require.NoError(t, err)

	_, err = db.Exec(`
		CREATE TABLE test_simple_pk (
			id INTEGER PRIMARY KEY
		);

		CREATE TABLE test_composite_pk (
			order_id INTEGER NOT NULL,
			product_id INTEGER NOT NULL,
			PRIMARY KEY (order_id, product_id)
		);

		CREATE TABLE test_named_pk (
			id INTEGER,
			CONSTRAINT pk_test_named PRIMARY KEY (id)
		);
	`)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err := db.Exec(`
			DROP TABLE IF EXISTS test_simple_pk;
			DROP TABLE IF EXISTS test_composite_pk;
			DROP TABLE IF EXISTS test_named_pk;
		`)
		require.NoError(t, err)
	})

	s := New()
	tableSingle := &schema.Table{
		Name: "test_simple_pk",
	}
	err = s.InspectPrimaryKey(db, tableSingle)
	require.NoError(t, err)
	require.Equal(t, &schema.PrimaryKey{
		Name:    "test_simple_pk_pkey",
		Columns: []string{"id"},
	}, tableSingle.PrimaryKey)

	tableComposite := &schema.Table{
		Name: "test_composite_pk",
	}
	err = s.InspectPrimaryKey(db, tableComposite)
	require.NoError(t, err)
	require.Equal(t, &schema.PrimaryKey{
		Name:    "test_composite_pk_pkey",
		Columns: []string{"order_id", "product_id"},
	}, tableComposite.PrimaryKey)

	tableNamed := &schema.Table{
		Name: "test_named_pk",
	}
	err = s.InspectPrimaryKey(db, tableNamed)
	require.NoError(t, err)
	require.Equal(t, &schema.PrimaryKey{
		Name:    "test_named_pk_pkey",
		Columns: []string{"id"},
	}, tableNamed.PrimaryKey)
}
