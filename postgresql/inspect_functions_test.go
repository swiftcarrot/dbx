package postgresql

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/swiftcarrot/dbx/internal/testutil"
	"github.com/swiftcarrot/dbx/schema"
)

func TestInspectFunctions(t *testing.T) {
	db, err := testutil.GetPGTestConn()
	require.NoError(t, err)

	_, err = db.Exec(`
		-- Simple function
		CREATE OR REPLACE FUNCTION test_add(a integer, b integer)
		RETURNS integer AS $$
		BEGIN
			RETURN a + b;
		END;
		$$ LANGUAGE plpgsql IMMUTABLE;

		-- Function with default parameter
		CREATE OR REPLACE FUNCTION test_multiply(a integer, b integer DEFAULT 2)
		RETURNS integer AS $$
		BEGIN
			RETURN a * b;
		END;
		$$ LANGUAGE plpgsql STABLE;

		-- Security definer function
		CREATE OR REPLACE FUNCTION test_secure_func()
		RETURNS text AS $$
		BEGIN
			RETURN 'secure';
		END;
		$$ LANGUAGE plpgsql SECURITY DEFINER;
	`)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err := db.Exec(`
			DROP FUNCTION IF EXISTS test_add(integer, integer);
			DROP FUNCTION IF EXISTS test_multiply(integer, integer);
			DROP FUNCTION IF EXISTS test_secure_func();
		`)
		require.NoError(t, err)
	})

	pg := New()
	s := schema.NewSchema()
	err = pg.InspectFunctions(db, s)
	require.NoError(t, err)
	require.Equal(t, []*schema.Function{
		{
			Schema: "public",
			Name:   "test_add",
			Arguments: []schema.FunctionArg{
				{
					Name: "a",
					Type: "integer",
					Mode: "IN",
				},
				{
					Name: "b",
					Type: "integer",
					Mode: "IN",
				},
			},
			Returns:    "integer",
			Language:   "plpgsql",
			Body:       "\n\t\tBEGIN\n\t\t\tRETURN a + b;\n\t\tEND;\n\t\t",
			Volatility: "IMMUTABLE",
			Security:   "INVOKER",
			Cost:       100,
		},
		{
			Schema: "public",
			Name:   "test_multiply",
			Arguments: []schema.FunctionArg{
				{
					Name: "a",
					Type: "integer",
					Mode: "IN",
				},
				{
					Name:    "b",
					Type:    "integer",
					Mode:    "IN",
					Default: "2",
				},
			},
			Returns:    "integer",
			Language:   "plpgsql",
			Body:       "\n\t\tBEGIN\n\t\t\tRETURN a * b;\n\t\tEND;\n\t\t",
			Volatility: "STABLE",
			Security:   "INVOKER",
			Cost:       100,
		},
		{
			Schema:     "public",
			Name:       "test_secure_func",
			Returns:    "text",
			Language:   "plpgsql",
			Body:       "\n\t\tBEGIN\n\t\t\tRETURN 'secure';\n\t\tEND;\n\t\t",
			Volatility: "VOLATILE",
			Security:   "DEFINER",
			Cost:       100,
		},
	}, s.Functions)
}
