package mysql

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/swiftcarrot/dbx/internal/testutil"
	"github.com/swiftcarrot/dbx/schema"
)

func TestInspectPrimaryKeys(t *testing.T) {
	db, err := testutil.GetMySQLTestConn()
	require.NoError(t, err)
	defer db.Close()

	setupSQL := `
		CREATE TABLE single_pk (
			id INT PRIMARY KEY,
			name VARCHAR(255)
		);

		CREATE TABLE composite_pk (
			id1 INT,
			id2 INT,
			data VARCHAR(255),
			PRIMARY KEY (id1, id2)
		);
	`
	_, err = db.Exec(setupSQL)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err := db.Exec(`
			DROP TABLE IF EXISTS single_pk;
			DROP TABLE IF EXISTS composite_pk;
		`)
		require.NoError(t, err)
	})
	my := New()

	singleTable := &schema.Table{
		Name: "single_pk",
		Columns: []*schema.Column{
			{Name: "id", Type: "int"},
			{Name: "name", Type: "varchar"},
		},
	}

	err = my.InspectPrimaryKey(db, singleTable)
	require.NoError(t, err)
	require.Equal(t, &schema.PrimaryKey{
		Columns: []string{"id"},
	}, singleTable.PrimaryKey)

	compositeTable := &schema.Table{
		Name: "composite_pk",
		Columns: []*schema.Column{
			{Name: "id1", Type: "int"},
			{Name: "id2", Type: "int"},
			{Name: "data", Type: "varchar"},
		},
	}

	err = my.InspectPrimaryKey(db, compositeTable)
	require.NoError(t, err)
	require.Equal(t, &schema.PrimaryKey{
		Columns: []string{"id1", "id2"},
	}, compositeTable.PrimaryKey)
}
