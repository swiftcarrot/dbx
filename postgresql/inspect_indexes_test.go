package postgresql

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/swiftcarrot/dbx/internal/testutil"
	"github.com/swiftcarrot/dbx/schema"
)

func TestInspectIndexes(t *testing.T) {
	db, err := testutil.GetPGTestConn()
	require.NoError(t, err)

	_, err = db.Exec(`
		CREATE TABLE test_indexes (
			id serial PRIMARY KEY,
			username varchar(50) NOT NULL,
			email varchar(100) NOT NULL,
			first_name varchar(50),
			last_name varchar(50),
			created_at timestamp NOT NULL
		);

		-- Regular index
		CREATE INDEX idx_test_username ON test_indexes (username);

		-- Unique index
		CREATE UNIQUE INDEX idx_test_email ON test_indexes (email);

		-- Multi-column index
		CREATE INDEX idx_test_name ON test_indexes (first_name, last_name);

		-- Expression index
		-- CREATE INDEX idx_test_created_year ON test_indexes (EXTRACT(YEAR FROM created_at));
	`)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, err = db.Exec(`DROP TABLE IF EXISTS test_indexes`)
		require.NoError(t, err)
	})

	pg := New()
	table := &schema.Table{
		Name:   "test_indexes",
		Schema: "public",
	}
	err = pg.InspectIndexes(db, table)
	require.NoError(t, err)

	require.Equal(t, []*schema.Index{
		{Name: "idx_test_email", Columns: []string{"email"}, Unique: true},
		{Name: "idx_test_name", Columns: []string{"first_name", "last_name"}},
		{Name: "idx_test_username", Unique: false, Columns: []string{"username"}},
		// TODO: functional index support
		// {Name: "idx_test_created_year", Unique: false, Columns: []string{"EXTRACT(YEAR FROM created_at)"}},
	}, table.Indexes)
}
