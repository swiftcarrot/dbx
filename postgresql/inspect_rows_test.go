package postgresql

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/swiftcarrot/dbx/internal/testutil"
)

func TestInspectRows(t *testing.T) {
	db, err := testutil.GetPGTestConn()
	require.NoError(t, err)

	// Create a test table with different data types
	_, err = db.Exec(`
		CREATE TABLE rows_test_table (
			id serial PRIMARY KEY,
			name varchar(50) NOT NULL,
			age integer,
			active boolean DEFAULT true,
			score numeric(5,2),
			tags jsonb,
			created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		-- Insert test data
		INSERT INTO rows_test_table (name, age, active, score, tags) VALUES
			('Alice', 25, true, 92.50, '["student", "developer"]'),
			('Bob', 30, false, 85.75, '{"role": "admin", "level": 3}'),
			('Charlie', 22, true, 78.25, NULL);
	`)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, err := db.Exec(`DROP TABLE IF EXISTS rows_test_table`)
		require.NoError(t, err)
	})

	// Test basic row inspection
	pg := New()
	rows, err := pg.InspectRows(db, "rows_test_table", 10, "")
	require.NoError(t, err)
	require.Len(t, rows, 3)

	// Verify first row
	require.Equal(t, int64(1), rows[0]["id"])
	require.Equal(t, "Alice", rows[0]["name"])
	require.Equal(t, int64(25), rows[0]["age"])
	require.Equal(t, true, rows[0]["active"])

	// Verify second row
	require.Equal(t, int64(2), rows[1]["id"])
	require.Equal(t, "Bob", rows[1]["name"])
	require.Equal(t, int64(30), rows[1]["age"])
	require.Equal(t, false, rows[1]["active"])

	// Test with WHERE clause
	filteredRows, err := pg.InspectRows(db, "rows_test_table", 10, "age > $1", 25)
	require.NoError(t, err)
	require.Len(t, filteredRows, 1)
	require.Equal(t, "Bob", filteredRows[0]["name"])

	// Test with LIMIT
	limitedRows, err := pg.InspectRows(db, "rows_test_table", 2, "")
	require.NoError(t, err)
	require.Len(t, limitedRows, 2)
}
