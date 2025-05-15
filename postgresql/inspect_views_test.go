package postgresql

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/swiftcarrot/dbx/internal/testutil"
	"github.com/swiftcarrot/dbx/schema"
)

func TestInspectViews(t *testing.T) {
	db, err := testutil.GetPGTestConn()
	require.NoError(t, err)

	def1 := " SELECT id,\n    name,\n    price\n   FROM view_test_items;"
	def2 := " SELECT id,\n    item_id\n   FROM view_test_orders;"
	_, err = db.Exec(`
		CREATE TABLE view_test_items (
			id serial PRIMARY KEY,
			name varchar(50) NOT NULL,
			price numeric(10,2) NOT NULL,
			created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE view_test_orders (
			id serial PRIMARY KEY,
			item_id integer NOT NULL REFERENCES view_test_items(id),
			quantity integer NOT NULL,
			created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		-- Simple view
		CREATE VIEW view_test_simple AS
		` + def1 + `;

		-- View with joins
		CREATE VIEW view_test_orders_detail AS
		` + def2 + `;
	`)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err := db.Exec(`
			DROP VIEW IF EXISTS view_test_orders_detail;
			DROP VIEW IF EXISTS view_test_simple;
			DROP TABLE IF EXISTS view_test_orders;
			DROP TABLE IF EXISTS view_test_items;
		`)
		require.NoError(t, err)
	})

	pg := New()
	s := schema.NewSchema()
	err = pg.InspectViews(db, s)
	require.NoError(t, err)
	require.Equal(t, []*schema.View{
		{
			Schema:     "public",
			Name:       "view_test_orders_detail",
			Definition: def2,
			Columns:    []string{"id", "item_id"},
			Options:    []string{},
		},
		{
			Schema:     "public",
			Name:       "view_test_simple",
			Definition: def1,
			Columns:    []string{"id", "name", "price"},
			Options:    []string{},
		},
	}, s.Views)
}
