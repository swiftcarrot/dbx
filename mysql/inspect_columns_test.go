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

	_, err = db.Exec(`
		CREATE TABLE test_columns (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(50) NOT NULL,
			description TEXT NULL,
			age INTEGER DEFAULT 18,
			rating DECIMAL(3,1) NOT NULL DEFAULT 5.0,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`)
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

	require.Equal(t, []*schema.Column{
		{Name: "id", Type: &schema.IntegerType{}, Nullable: false, AutoIncrement: true},
		{Name: "name", Type: &schema.VarcharType{Length: 50}, Nullable: false},
		{Name: "description", Type: &schema.TextType{}, Nullable: true},
		{Name: "age", Type: &schema.IntegerType{}, Nullable: true, Default: "18"},
		{Name: "rating", Type: &schema.DecimalType{Precision: 3, Scale: 1}, Nullable: false, Default: "5.0"},
		{Name: "created_at", Type: &schema.TimestampType{}, Nullable: false, Default: "CURRENT_TIMESTAMP"},
	}, table.Columns)
}
