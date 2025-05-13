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
		CREATE TABLE test_trigger_users (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE test_trigger_audit (
			id INTEGER PRIMARY KEY,
			table_name TEXT NOT NULL,
			operation TEXT NOT NULL,
			timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TRIGGER trg_users_update
		AFTER UPDATE ON test_trigger_users
		FOR EACH ROW
		BEGIN
			UPDATE test_trigger_users SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
		END;

		CREATE TRIGGER trg_users_audit
		AFTER INSERT OR UPDATE OR DELETE ON test_trigger_users
		FOR EACH ROW
		WHEN (OLD.name IS DISTINCT FROM NEW.name OR OLD.email IS DISTINCT FROM NEW.email)
		BEGIN
			INSERT INTO test_trigger_audit (table_name, operation, timestamp)
			VALUES ('test_trigger_users', CASE
				WHEN OLD.id IS NULL THEN 'INSERT'
				WHEN NEW.id IS NULL THEN 'DELETE'
				ELSE 'UPDATE'
			END, CURRENT_TIMESTAMP);
		END;
	`)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err := db.Exec(`
			DROP TRIGGER IF EXISTS trg_users_update;
			DROP TRIGGER IF EXISTS trg_users_audit;
			DROP TABLE IF EXISTS test_trigger_audit;
			DROP TABLE IF EXISTS test_trigger_users;
		`)
		require.NoError(t, err)
	})

	s := New()
	schm := schema.NewSchema()
	err = s.InspectTriggers(db, schm)
	require.NoError(t, err)

	require.Equal(t, []*schema.Trigger{
		{
			Name:      "trg_users_audit",
			Table:     "test_trigger_users",
			Events:    []string{"INSERT", "UPDATE", "DELETE"},
			Timing:    "AFTER",
			ForEach:   "ROW",
			When:      "OLD.name IS DISTINCT FROM NEW.name OR OLD.email IS DISTINCT FROM NEW.email",
			Function:  "",
			Arguments: []string{},
		},
		{
			Name:      "trg_users_update",
			Table:     "test_trigger_users",
			Events:    []string{"UPDATE"},
			Timing:    "AFTER",
			ForEach:   "ROW",
			When:      "",
			Function:  "",
			Arguments: []string{},
		},
	}, schm.Triggers)
}
