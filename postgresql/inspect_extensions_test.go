package postgresql

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/swiftcarrot/dbx/internal/testutil"
	"github.com/swiftcarrot/dbx/schema"
)

func TestInspectExtensions(t *testing.T) {
	db, err := testutil.GetPGTestConn()
	require.NoError(t, err)

	_, err = db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err := db.Exec(`DROP EXTENSION IF EXISTS "uuid-ossp"`)
		require.NoError(t, err)
	})

	pg := New()
	s := schema.NewSchema()
	err = pg.InspectExtensions(db, s)
	require.NoError(t, err)
	require.Contains(t, s.Extensions, "uuid-ossp")
}
