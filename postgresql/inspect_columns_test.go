package postgresql

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/swiftcarrot/dbx/internal/testutil"
	"github.com/swiftcarrot/dbx/schema"
)

func TestInspectColumns(t *testing.T) {
	db, err := testutil.GetPGTestConn()
	require.NoError(t, err)

	_, err = db.Exec(`
		CREATE TABLE test_columns (
			id serial PRIMARY KEY,
			name varchar(50) NOT NULL,
			description text NULL,
			age integer DEFAULT 18,
			rating numeric(3,1) NOT NULL DEFAULT 5.0,
			created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		COMMENT ON COLUMN test_columns.description IS 'Description of the entity';
	`)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err := db.Exec(`DROP TABLE IF EXISTS test_columns`)
		require.NoError(t, err)
	})

	pg := New()
	table := &schema.Table{
		Schema: "public",
		Name:   "test_columns",
	}

	err = pg.InspectColumns(db, table)
	require.NoError(t, err)
	require.Equal(t, []*schema.Column{
		{Name: "id", Type: &schema.IntegerType{}, Nullable: false, Default: "nextval('test_columns_id_seq'::regclass)"},
		{Name: "name", Type: &schema.VarcharType{}, Nullable: false},
		{Name: "description", Type: &schema.TextType{}, Nullable: true, Comment: "Description of the entity"},
		{Name: "age", Type: &schema.IntegerType{}, Nullable: true, Default: "18"},
		{Name: "rating", Type: &schema.DecimalType{Precision: 3, Scale: 1}, Nullable: false, Default: "5.0"},
		{Name: "created_at", Type: &schema.TimestampType{}, Nullable: false, Default: "CURRENT_TIMESTAMP"},
	}, table.Columns)
}
