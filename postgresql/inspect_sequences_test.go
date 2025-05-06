package postgresql

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/swiftcarrot/dbx/internal/testutil"
	"github.com/swiftcarrot/dbx/schema"
)

func TestInspectSequences(t *testing.T) {
	db, err := testutil.GetPGTestConn()
	require.NoError(t, err)

	_, err = db.Exec(`
		CREATE SEQUENCE test_sequence_1
		INCREMENT 2
		START 100
		MINVALUE 50
		MAXVALUE 1000
		CACHE 5;

		CREATE SEQUENCE test_sequence_2
		INCREMENT 5
		START 1
		CYCLE;
	`)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err := db.Exec(`
			DROP SEQUENCE IF EXISTS test_sequence_1;
			DROP SEQUENCE IF EXISTS test_sequence_2;
		`)
		require.NoError(t, err)
	})

	pg := New()
	s := schema.NewSchema()
	err = pg.InspectSequences(db, s)
	require.NoError(t, err)
	require.Equal(t, []*schema.Sequence{
		{
			Name:      "test_sequence_1",
			Start:     100,
			Increment: 2,
			MinValue:  50,
			MaxValue:  1000,
			Cache:     5,
			Cycle:     false,
		},
		{
			Name:      "test_sequence_2",
			Start:     1,
			Increment: 5,
			MinValue:  1,
			MaxValue:  9223372036854775807,
			Cache:     1,
			Cycle:     true,
		},
	}, s.Sequences)
}
