package mysql

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/swiftcarrot/dbx/internal/testutil"
)

func TestInspectTables(t *testing.T) {
	db, err := testutil.GetMySQLTestConn()
	require.NoError(t, err)
	defer db.Close()

	setupSQL := `
		CREATE TABLE users (
			id INT PRIMARY KEY
		);

		CREATE TABLE posts (
			id INT PRIMARY KEY
		);

		CREATE TABLE comments (
			id INT PRIMARY KEY
		);
	`
	_, err = db.Exec(setupSQL)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, err := db.Exec(`
			DROP TABLE IF EXISTS users;
			DROP TABLE IF EXISTS posts;
			DROP TABLE IF EXISTS comments;
		`)
		require.NoError(t, err)
	})

	my := New()

	tables, err := my.InspectTables(db)
	require.NoError(t, err)

	expectedTables := []string{"users", "posts", "comments"}
	require.Equal(t, expectedTables, tables)
}
