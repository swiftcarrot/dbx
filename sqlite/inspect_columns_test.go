package sqlite

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/swiftcarrot/dbx/internal/testutil"
	"github.com/swiftcarrot/dbx/schema"
)

func TestInspectColumns(t *testing.T) {
	db, err := testutil.GetSQLiteTestConn()
	require.NoError(t, err)

	_, err = db.Exec(`
		CREATE TABLE test_columns (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT NULL,
			age INTEGER DEFAULT 18,
		-- 	rating NUMERIC(3,1) NOT NULL DEFAULT 5.0,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err := db.Exec(`DROP TABLE IF EXISTS test_columns`)
		require.NoError(t, err)
	})

	s := New()
	table := &schema.Table{
		Name: "test_columns",
	}

	err = s.InspectColumns(db, table)
	require.NoError(t, err)
	require.Equal(t, []*schema.Column{
		{
			Name:     "id",
			Type:     &IntegerType{},
			Nullable: true,
		},
		{
			Name:     "name",
			Type:     &TextType{},
			Nullable: false,
		},
		{
			Name:     "description",
			Type:     &TextType{},
			Nullable: true,
		},
		{
			Name:     "age",
			Type:     &IntegerType{},
			Nullable: true,
			Default:  "18",
		},
		{
			Name:      "rating",
			Type:      &schema.DecimalType{Precision: 3, Scale: 1},
			Precision: 3,
			Scale:     1,
			Nullable:  false,
			Default:   "5.0",
		},
		{
			Name:     "created_at",
			Type:     &schema.TimestampType{},
			Nullable: false,
			Default:  "CURRENT_TIMESTAMP",
		},
	}, table.Columns)
}
