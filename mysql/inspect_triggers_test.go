package mysql

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/swiftcarrot/dbx/internal/testutil"
	"github.com/swiftcarrot/dbx/schema"
)

func TestInspectTriggers(t *testing.T) {
	db, err := testutil.GetMySQLTestConn()
	require.NoError(t, err)
	defer db.Close()

	setupSQL := `
		CREATE TABLE users (
			id INT PRIMARY KEY,
			name VARCHAR(255)
		);

		CREATE TABLE posts (
			id INT PRIMARY KEY,
			title VARCHAR(255),
			content TEXT,
			updated_at TIMESTAMP
		);

		CREATE TABLE audit_log (
			id INT AUTO_INCREMENT PRIMARY KEY,
			table_name VARCHAR(255),
			action VARCHAR(50),
			user_id INT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TRIGGER after_insert_users
		AFTER INSERT ON users
		FOR EACH ROW
		INSERT INTO audit_log (table_name, action, user_id)
		VALUES ('users', 'INSERT', NEW.id);

		CREATE TRIGGER before_update_posts
		BEFORE UPDATE ON posts
		FOR EACH ROW
		SET NEW.updated_at = NOW();
	`
	_, err = db.Exec(setupSQL)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, err := db.Exec(`
			DROP TRIGGER IF EXISTS after_insert_users;
			DROP TRIGGER IF EXISTS before_update_posts;
			DROP TABLE IF EXISTS audit_log;
			DROP TABLE IF EXISTS users;
			DROP TABLE IF EXISTS posts;
		`)
		require.NoError(t, err)
	})

	my := New()

	s := &schema.Schema{}

	err = my.InspectTriggers(db, s)
	require.NoError(t, err)

	expectedTriggers := []*schema.Trigger{
		{
			Name:    "after_insert_users",
			Table:   "users",
			Events:  []string{"INSERT"},
			Timing:  "AFTER",
			ForEach: "ROW",
		},
		{
			Name:    "before_update_posts",
			Table:   "posts",
			Events:  []string{"UPDATE"},
			Timing:  "BEFORE",
			ForEach: "ROW",
		},
	}
	require.Equal(t, expectedTriggers, s.Triggers)
}
