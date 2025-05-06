package postgresql

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/swiftcarrot/dbx/internal/testutil"
)

func TestInspectSchemas(t *testing.T) {
	db, err := testutil.GetPGTestConn()
	require.NoError(t, err)

	_, err = db.Exec(`CREATE SCHEMA IF NOT EXISTS test_schema`)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err = db.Exec(`DROP SCHEMA IF EXISTS test_schema`)
		require.NoError(t, err)
	})

	pg := New()
	schemas, err := pg.InspectSchemas(db)
	require.NoError(t, err)

	require.Contains(t, schemas, "public")
	require.Contains(t, schemas, "test_schema")
}
