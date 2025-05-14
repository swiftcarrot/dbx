package mysql

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/swiftcarrot/dbx/internal/testutil"
	"github.com/swiftcarrot/dbx/schema"
)

func TestInspectViews(t *testing.T) {
	db, err := testutil.GetMySQLTestConn()
	require.NoError(t, err)

	_, err = db.Exec(`
		CREATE TABLE users (
			id INT PRIMARY KEY,
			name VARCHAR(255),
			email VARCHAR(255),
			active BOOLEAN
		);

		CREATE VIEW active_users AS
		SELECT * FROM users WHERE active = 1;
	`)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, err := db.Exec(`
			DROP VIEW IF EXISTS active_users;
			DROP TABLE IF EXISTS users;
		`)
		require.NoError(t, err)
	})

	my := New()
	s := &schema.Schema{}
	err = my.InspectViews(db, s)
	require.NoError(t, err)

	require.Equal(t, []*schema.View{
		{
			Name:       "active_users",
			Columns:    []string{"id", "name", "email", "active"},
			Definition: "select `dbx_test`.`users`.`id` AS `id`,`dbx_test`.`users`.`name` AS `name`,`dbx_test`.`users`.`email` AS `email`,`dbx_test`.`users`.`active` AS `active` from `dbx_test`.`users` where (`dbx_test`.`users`.`active` = 1)",
		},
	}, s.Views)
}
