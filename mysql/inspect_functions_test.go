package mysql

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/swiftcarrot/dbx/internal/testutil"
	"github.com/swiftcarrot/dbx/schema"
)

func TestInspectFunctions(t *testing.T) {
	db, err := testutil.GetMySQLTestConn()
	require.NoError(t, err)

	t.Cleanup(func() {
		_, err := db.Exec("DROP FUNCTION IF EXISTS get_user_name;")
		require.NoError(t, err)
	})

	_, err = db.Exec(`
		CREATE FUNCTION get_user_name(user_id INT)
		RETURNS VARCHAR(255) DETERMINISTIC
		RETURN CONCAT('user', user_id);
	`)
	require.NoError(t, err)

	s := &schema.Schema{}
	err = New().InspectFunctions(db, s)
	require.NoError(t, err)
	require.Equal(t, []*schema.Function{
		{
			Name:       "get_user_name",
			Arguments:  []schema.FunctionArg{{Name: "user_id", Type: "int(10)"}},
			Returns:    "varchar",
			Body:       "RETURN CONCAT('user', user_id)",
			Volatility: "IMMUTABLE",
		},
	}, s.Functions)
}
