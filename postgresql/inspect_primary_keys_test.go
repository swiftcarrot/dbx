package postgresql

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/swiftcarrot/dbx/internal/testutil"
	"github.com/swiftcarrot/dbx/schema"
)

func TestInspectPrimaryKey(t *testing.T) {
	db, err := testutil.GetPGTestConn()
	require.NoError(t, err)

	_, err = db.Exec(`
		CREATE TABLE test_simple_pk (
			id serial PRIMARY KEY
		);

		CREATE TABLE test_composite_pk (
			order_id integer NOT NULL,
			product_id integer NOT NULL,
			PRIMARY KEY (order_id, product_id)
		);

		CREATE TABLE test_named_pk (
			id serial,
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

	pg := New()
	tableSingle := &schema.Table{
		Name:   "test_simple_pk",
		Schema: "public",
	}
	err = pg.InspectPrimaryKey(db, tableSingle)
	require.NoError(t, err)
	require.Equal(t, &schema.PrimaryKey{
		Name:    "test_simple_pk_pkey",
		Columns: []string{"id"},
	}, tableSingle.PrimaryKey)

	tableComposite := &schema.Table{
		Name:   "test_composite_pk",
		Schema: "public",
	}
	err = pg.InspectPrimaryKey(db, tableComposite)
	require.NoError(t, err)
	require.Equal(t, &schema.PrimaryKey{
		Name:    "test_composite_pk_pkey",
		Columns: []string{"order_id", "product_id"},
	}, tableComposite.PrimaryKey)

	tableNamed := &schema.Table{
		Name:   "test_named_pk",
		Schema: "public",
	}
	err = pg.InspectPrimaryKey(db, tableNamed)
	require.NoError(t, err)
	require.Equal(t, &schema.PrimaryKey{
		Name:    "pk_test_named",
		Columns: []string{"id"},
	}, tableNamed.PrimaryKey)
}
