package postgresql

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/swiftcarrot/dbx/internal/testutil"
	"github.com/swiftcarrot/dbx/schema"
)

func TestInspectRowPolicies(t *testing.T) {
	db, err := testutil.GetPGTestConn()
	require.NoError(t, err)

	_, err = db.Exec(`
		CREATE TABLE policy_test_items (
			id serial PRIMARY KEY,
			name varchar(50) NOT NULL,
			owner varchar(50) NOT NULL,
			price numeric(10,2) NOT NULL,
			created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		ALTER TABLE policy_test_items ENABLE ROW LEVEL SECURITY;

		-- Create a policy that only allows users to see their own items
		CREATE POLICY owner_policy ON policy_test_items
			FOR SELECT
			USING (owner = current_user);

		-- Create a policy for updates (without specific role)
		CREATE POLICY admin_update_policy ON policy_test_items
			FOR UPDATE
			USING (true);

		-- Create a restrictive policy
		CREATE POLICY restrictive_delete_policy ON policy_test_items
			AS RESTRICTIVE
			FOR DELETE
			USING (price <= 100.00);
	`)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err := db.Exec(`
			DROP TABLE IF EXISTS policy_test_items;
		`)
		require.NoError(t, err)
	})

	pg := New()
	s := schema.NewSchema()
	err = pg.InspectRowPolicies(db, s)
	require.NoError(t, err)

	require.Equal(t, []*schema.RowPolicy{
		{
			Schema:      "public",
			TableName:   "policy_test_items",
			PolicyName:  "admin_update_policy",
			CommandType: "UPDATE",
			Roles:       []string{"public"},
			UsingExpr:   "true",
			Permissive:  true,
		},
		{
			Schema:      "public",
			TableName:   "policy_test_items",
			PolicyName:  "owner_policy",
			CommandType: "SELECT",
			Roles:       []string{"public"},
			UsingExpr:   "((owner)::text = CURRENT_USER)",
			Permissive:  true,
		},
		{
			Schema:      "public",
			TableName:   "policy_test_items",
			PolicyName:  "restrictive_delete_policy",
			CommandType: "DELETE",
			Roles:       []string{"public"},
			UsingExpr:   "(price <= 100.00)",
			Permissive:  false,
		},
	}, s.RowPolicies)
}
