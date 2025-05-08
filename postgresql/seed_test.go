package postgresql

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/swiftcarrot/dbx/internal/testutil"
	"github.com/swiftcarrot/dbx/seed"
)

func TestImportFromCSV(t *testing.T) {
	db, err := testutil.GetPGTestConn()
	require.NoError(t, err)

	_, err = db.Exec(`
		CREATE TABLE csv_import_test (
			id integer,
			name varchar(50),
			email varchar(100),
			active boolean,
			score numeric(5,2)
		)
	`)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, err := db.Exec(`DROP TABLE IF EXISTS csv_import_test`)
		require.NoError(t, err)
	})

	t.Run("BasicImport", func(t *testing.T) {
		csvData := `id,name,email,active,score
1,John Doe,john@example.com,true,85.50
2,Jane Smith,jane@example.com,false,92.75
3,Bob Johnson,bob@example.com,true,78.25`

		reader := strings.NewReader(csvData)
		pg := New()

		err = pg.ImportFromCSV(db, "csv_import_test", "public", reader, &seed.CSVImportOptions{
			Delimiter: ",",
			Header:    true,
			Columns:   []string{"id", "name", "email", "active", "score"},
		})
		require.NoError(t, err)

		rows, err := db.Query("SELECT * FROM csv_import_test ORDER BY id")
		require.NoError(t, err)
		defer rows.Close()

		var results []struct {
			ID     int
			Name   string
			Email  string
			Active bool
			Score  float64
		}

		for rows.Next() {
			var r struct {
				ID     int
				Name   string
				Email  string
				Active bool
				Score  float64
			}
			err := rows.Scan(&r.ID, &r.Name, &r.Email, &r.Active, &r.Score)
			require.NoError(t, err)
			results = append(results, r)
		}

		require.Equal(t, []struct {
			ID     int
			Name   string
			Email  string
			Active bool
			Score  float64
		}{
			{ID: 1, Name: "John Doe", Email: "john@example.com", Active: true, Score: 85.50},
			{ID: 2, Name: "Jane Smith", Email: "jane@example.com", Active: false, Score: 92.75},
			{ID: 3, Name: "Bob Johnson", Email: "bob@example.com", Active: true, Score: 78.25},
		}, results)

		_, err = db.Exec("DELETE FROM csv_import_test")
		require.NoError(t, err)
	})

	t.Run("CustomDelimiterNoHeaders", func(t *testing.T) {
		csvData := `4|Alice Wilson|alice@example.com|true|91.40
5|Charlie Brown|charlie@example.com|false|68.30`

		reader := strings.NewReader(csvData)
		pg := New()

		err = pg.ImportFromCSV(db, "csv_import_test", "public", reader, &seed.CSVImportOptions{
			Delimiter: "|",
			Header:    false,
			Columns:   []string{"id", "name", "email", "active", "score"},
		})
		require.NoError(t, err)

		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM csv_import_test").Scan(&count)
		require.NoError(t, err)
		require.Equal(t, 2, count)

		var id int
		var name string
		var active bool
		err = db.QueryRow("SELECT id, name, active FROM csv_import_test WHERE email = 'charlie@example.com'").Scan(&id, &name, &active)
		require.NoError(t, err)
		require.Equal(t, 5, id)
		require.Equal(t, "Charlie Brown", name)
		require.Equal(t, false, active)
	})
}
