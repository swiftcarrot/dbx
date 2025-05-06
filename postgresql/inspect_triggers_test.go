package postgresql

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/swiftcarrot/dbx/internal/testutil"
	"github.com/swiftcarrot/dbx/schema"
)

func TestInspectTriggers(t *testing.T) {
	db, err := testutil.GetPGTestConn()
	require.NoError(t, err)

	_, err = db.Exec(`
		CREATE TABLE trigger_test_table (
			id serial PRIMARY KEY,
			name varchar(50),
			created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at timestamp
		);

		-- Create trigger functions
		CREATE OR REPLACE FUNCTION trigger_test_update_timestamp()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = CURRENT_TIMESTAMP;
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;

		CREATE OR REPLACE FUNCTION trigger_test_audit_log()
		RETURNS TRIGGER AS $$
		BEGIN
			RAISE NOTICE 'Audit: % on %', TG_OP, TG_TABLE_NAME;
			RETURN NULL;
		END;
		$$ LANGUAGE plpgsql;

		-- Create triggers
		CREATE TRIGGER trigger_test_before_update
		BEFORE UPDATE ON trigger_test_table
		FOR EACH ROW
		EXECUTE FUNCTION trigger_test_update_timestamp();

		CREATE TRIGGER trigger_test_after_insert
		AFTER INSERT ON trigger_test_table
		FOR EACH ROW
		EXECUTE FUNCTION trigger_test_audit_log();

		CREATE TRIGGER trigger_test_when_condition
		BEFORE UPDATE ON trigger_test_table
		FOR EACH ROW
		WHEN (OLD.name IS DISTINCT FROM NEW.name)
		EXECUTE FUNCTION trigger_test_audit_log();
	`)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err := db.Exec(`
			DROP TRIGGER IF EXISTS trigger_test_before_update ON trigger_test_table;
			DROP TRIGGER IF EXISTS trigger_test_after_insert ON trigger_test_table;
			DROP TRIGGER IF EXISTS trigger_test_when_condition ON trigger_test_table;
			DROP FUNCTION IF EXISTS trigger_test_update_timestamp();
			DROP FUNCTION IF EXISTS trigger_test_audit_log();
			DROP TABLE IF EXISTS trigger_test_table;
		`)
		require.NoError(t, err)
	})

	pg := New()
	s := schema.NewSchema()
	err = pg.InspectTriggers(db, s)
	require.NoError(t, err)

	require.Equal(t, []*schema.Trigger{
		{
			Schema:    "public",
			Name:      "trigger_test_after_insert",
			Table:     "trigger_test_table",
			Events:    []string{"INSERT"},
			Timing:    "AFTER",
			ForEach:   "ROW",
			Function:  "trigger_test_audit_log",
			Arguments: []string{},
		},
		{
			Schema:    "public",
			Name:      "trigger_test_before_update",
			Table:     "trigger_test_table",
			Events:    []string{"UPDATE"},
			Timing:    "BEFORE",
			ForEach:   "ROW",
			Function:  "trigger_test_update_timestamp",
			Arguments: []string{},
		},
		{
			Schema:    "public",
			Name:      "trigger_test_when_condition",
			Table:     "trigger_test_table",
			Events:    []string{"UPDATE"},
			Timing:    "BEFORE",
			ForEach:   "ROW",
			When:      "((old.name)::text IS DISTINCT FROM (new.name)::text)",
			Function:  "trigger_test_audit_log",
			Arguments: []string{},
		},
	}, s.Triggers)
}
