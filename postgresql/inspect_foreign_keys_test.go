package postgresql

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/swiftcarrot/dbx/internal/testutil"
	"github.com/swiftcarrot/dbx/schema"
)

func TestInspectForeignKeys(t *testing.T) {
	db, err := testutil.GetPGTestConn()
	require.NoError(t, err)

	_, err = db.Exec(`
		CREATE TABLE fk_parent (
			id serial PRIMARY KEY,
			name varchar(50) NOT NULL,
			UNIQUE (id, name)
		);

		CREATE TABLE fk_child (
			id serial PRIMARY KEY,
			parent_id integer NOT NULL,
			secondary_parent_id integer,
			CONSTRAINT fk_child_parent FOREIGN KEY (parent_id) REFERENCES fk_parent (id) ON DELETE CASCADE,
			CONSTRAINT fk_child_secondary FOREIGN KEY (secondary_parent_id) REFERENCES fk_parent (id) ON DELETE SET NULL
		);

		-- Table with composite foreign key
		CREATE TABLE fk_composite (
			id serial PRIMARY KEY,
			parent_id integer NOT NULL,
			parent_name varchar(50) NOT NULL,
			CONSTRAINT fk_composite_parent FOREIGN KEY (parent_id, parent_name)
			REFERENCES fk_parent (id, name) ON UPDATE CASCADE
		);
	`)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err = db.Exec(`
		DROP TABLE IF EXISTS fk_composite;
		DROP TABLE IF EXISTS fk_child;
		DROP TABLE IF EXISTS fk_parent;
	`)
		require.NoError(t, err)
	})

	pg := New()

	childTable := &schema.Table{
		Name:   "fk_child",
		Schema: "public",
	}
	err = pg.InspectForeignKeys(db, childTable)
	require.NoError(t, err)
	require.Equal(t, []*schema.ForeignKey{
		{
			Name:       "fk_child_parent",
			Columns:    []string{"parent_id"},
			RefTable:   "fk_parent",
			RefColumns: []string{"id"},
			OnDelete:   "CASCADE",
			OnUpdate:   "",
		},
		{
			Name:       "fk_child_secondary",
			Columns:    []string{"secondary_parent_id"},
			RefTable:   "fk_parent",
			RefColumns: []string{"id"},
			OnDelete:   "SET NULL",
			OnUpdate:   "",
		},
	}, childTable.ForeignKeys)

	// TODO: Uncomment and fix the composite foreign key test
	// compositeTable := &schema.Table{
	// 	Name:   "fk_composite",
	// 	Schema: "public",
	// }
	// err = pg.InspectForeignKeys(db, compositeTable)
	// require.NoError(t, err)
	// require.Equal(t, []*schema.ForeignKey{
	// 	{
	// 		Name:       "fk_composite_parent",
	// 		Columns:    []string{"parent_id", "parent_name"},
	// 		RefTable:   "fk_parent",
	// 		RefColumns: []string{"id", "name"},
	// 		OnDelete:   "",
	// 		OnUpdate:   "CASCADE",
	// 	},
	// }, compositeTable.ForeignKeys)
}
