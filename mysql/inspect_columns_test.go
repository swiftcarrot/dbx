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
			name VARCHAR(255) NOT NULL,
			description TEXT,
			amount DECIMAL(10,2) DEFAULT 0.00,
			is_active BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
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
		{Name: "name", Type: &schema.VarcharType{Length: 255}, Nullable: false},
		{Name: "description", Type: &schema.TextType{}, Nullable: true},
		{Name: "amount", Type: &schema.DecimalType{Precision: 10, Scale: 2}, Precision: 10, Scale: 2, Nullable: true, Default: "0.00"},
		{Name: "is_active", Type: &schema.IntegerType{}, Nullable: true, Default: "1"},
		{Name: "created_at", Type: &schema.TimestampType{}, Nullable: true, Default: "CURRENT_TIMESTAMP"},
	}, table.Columns)
}
