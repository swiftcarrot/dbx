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
	defer db.Close()

	setupSQL := `
		CREATE FUNCTION get_user_name(user_id INT)
		RETURNS VARCHAR(255) DETERMINISTIC
		RETURN CONCAT('user', user_id);

		CREATE FUNCTION calculate_total(price DECIMAL(10,2), quantity INT)
		RETURNS DECIMAL(10,2) NOT DETERMINISTIC
		RETURN price * quantity;
	`
	_, err = db.Exec(setupSQL)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err := db.Exec("DROP FUNCTION IF EXISTS get_user_name; DROP FUNCTION IF EXISTS calculate_total;")
		require.NoError(t, err)
	})

	my := New()

	s := &schema.Schema{}

	err = my.InspectFunctions(db, s)
	require.NoError(t, err)

	expectedFunctions := []*schema.Function{
		{
			Name:       "calculate_total",
			Arguments:  []schema.FunctionArg{{Name: "price", Type: "decimal"}, {Name: "quantity", Type: "int"}},
			Returns:    "decimal",
			Volatility: "STABLE",
		},
		{
			Name:       "get_user_name",
			Arguments:  []schema.FunctionArg{{Name: "user_id", Type: "int"}},
			Returns:    "varchar",
			Volatility: "IMMUTABLE",
		},
	}
	require.Equal(t, expectedFunctions, s.Functions)
}
