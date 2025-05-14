package sqlite

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/swiftcarrot/dbx/internal/testutil"
	"github.com/swiftcarrot/dbx/schema"
)

func TestInspectTriggers(t *testing.T) {
	db, err := testutil.GetSQLiteTestConn()
	require.NoError(t, err)

	_, err = db.Exec(`
		CREATE TABLE test_users (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT NOT NULL,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TRIGGER trg_users_update
		AFTER UPDATE ON test_users
		FOR EACH ROW
		BEGIN
			UPDATE test_users SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
		END;
	`)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err := db.Exec(`
			DROP TRIGGER IF EXISTS trg_users_update;
			DROP TABLE IF EXISTS test_users;
		`)
		require.NoError(t, err)
	})

	s := New()
	sch := schema.NewSchema()
	err = s.InspectTriggers(db, sch)
	require.NoError(t, err)

	require.Equal(t, []*schema.Trigger{
		{
			Name:      "trg_users_update",
			Table:     "test_users",
			Events:    []string{"UPDATE"},
			Timing:    "AFTER",
			ForEach:   "ROW",
			When:      "",
			Function:  "",
			Arguments: []string{},
		},
	}, sch.Triggers)
}
