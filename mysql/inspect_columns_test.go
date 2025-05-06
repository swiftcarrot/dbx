package mysql

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/swiftcarrot/dbx/internal/testutil"
	"github.com/swiftcarrot/dbx/schema"
)

func TestInspectColumns(t *testing.T) {
	db, err := testutil.GetMySQLTestConn()
	require.NoError(t, err)
	defer db.Close()

	setupSQL := `
		CREATE TABLE test_columns (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			amount DECIMAL(10,2) DEFAULT 0.00,
			is_active BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`
	_, err = db.Exec(setupSQL)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err := db.Exec("DROP TABLE IF EXISTS test_columns;")
		require.NoError(t, err)
	})

	my := New()

	table := &schema.Table{
		Name: "test_columns",
	}

	err = my.InspectColumns(db, table)
	require.NoError(t, err)

	expectedColumns := []*schema.Column{
		{Name: "id", Type: "int", Nullable: false, PrimaryKey: true, AutoIncrement: true},
		{Name: "name", Type: "varchar", Length: 255, Nullable: false},
		{Name: "description", Type: "text", Nullable: true},
		{Name: "amount", Type: "decimal", Precision: 10, Scale: 2, Nullable: true, Default: "0.00"},
		{Name: "is_active", Type: "tinyint", Nullable: true, Default: "1"},
		{Name: "created_at", Type: "timestamp", Nullable: true, Default: "CURRENT_TIMESTAMP"},
	}

	require.Equal(t, expectedColumns, table.Columns)
}
