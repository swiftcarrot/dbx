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

		CREATE TABLE test_orders (
			id INTEGER PRIMARY KEY,
			customer_name TEXT NOT NULL
		);

		CREATE TABLE test_order_items (
			order_id INTEGER NOT NULL,
			product_id INTEGER NOT NULL,
			quantity INTEGER NOT NULL,
			PRIMARY KEY (order_id, product_id),
			FOREIGN KEY (order_id) REFERENCES test_orders(id),
			FOREIGN KEY (product_id) REFERENCES test_products(id)
		);

		CREATE VIEW view_test_simple AS
		SELECT id, name, price FROM test_products WHERE price > 0;

		CREATE VIEW view_test_orders_detail AS
		SELECT
			oi.order_id,
			p.name as item_name,
			oi.quantity,
			p.price,
			(oi.quantity * p.price) as total_price
		FROM test_order_items oi
		JOIN test_products p ON p.id = oi.product_id
		ORDER BY oi.order_id;
	`)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err := db.Exec(`
			DROP VIEW IF EXISTS view_test_simple;
			DROP VIEW IF EXISTS view_test_orders_detail;
			DROP TABLE IF EXISTS test_order_items;
			DROP TABLE IF EXISTS test_orders;
			DROP TABLE IF EXISTS test_products;
		`)
		require.NoError(t, err)
	})

	s := New()
	schm := schema.NewSchema()
	err = s.InspectViews(db, schm)
	require.NoError(t, err)

	def1 := "SELECT id, name, price FROM test_products WHERE price > 0"
	def2 := "SELECT oi.order_id, p.name as item_name, oi.quantity, p.price, (oi.quantity * p.price) as total_price FROM test_order_items oi JOIN test_products p ON p.id = oi.product_id ORDER BY oi.order_id"

	require.Equal(t, []*schema.View{
		{
			Name:       "view_test_orders_detail",
			Definition: def2,
			Columns:    []string{"order_id", "item_name", "quantity", "price", "total_price"},
		},
		{
			Name:       "view_test_simple",
			Definition: def1,
			Columns:    []string{"id", "name", "price"},
		},
	}, schm.Views)
}
