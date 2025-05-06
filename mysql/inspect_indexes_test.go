package mysql

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/swiftcarrot/dbx/internal/testutil"
	"github.com/swiftcarrot/dbx/schema"
)

func TestInspectIndexes(t *testing.T) {
	db, err := testutil.GetMySQLTestConn()
	require.NoError(t, err)
	defer db.Close()

	setupSQL := `
		CREATE TABLE test_indexes (
			id INT PRIMARY KEY,
			email VARCHAR(255),
			first_name VARCHAR(255),
			last_name VARCHAR(255),
			created_at TIMESTAMP,
			UNIQUE INDEX idx_email (email),
			INDEX idx_name (first_name, last_name),
			INDEX idx_created_at (created_at)
		);
	`
	_, err = db.Exec(setupSQL)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err := db.Exec("DROP TABLE IF EXISTS test_indexes;")
		require.NoError(t, err)
	})

	my := New()

	table := &schema.Table{
		Name: "test_indexes",
		Columns: []*schema.Column{
			{Name: "id", Type: "int"},
			{Name: "email", Type: "varchar"},
			{Name: "first_name", Type: "varchar"},
			{Name: "last_name", Type: "varchar"},
			{Name: "created_at", Type: "timestamp"},
		},
	}

	err = my.InspectIndexes(db, table)
	require.NoError(t, err)

	expectedIndexes := []*schema.Index{
		{Name: "idx_created_at", Columns: []string{"created_at"}, Unique: false},
		{Name: "idx_email", Columns: []string{"email"}, Unique: true},
		{Name: "idx_name", Columns: []string{"first_name", "last_name"}, Unique: false},
	}
	require.Equal(t, expectedIndexes, table.Indexes)
}
