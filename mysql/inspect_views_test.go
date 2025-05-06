package mysql

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/swiftcarrot/dbx/internal/testutil"
	"github.com/swiftcarrot/dbx/schema"
)

func TestInspectViews(t *testing.T) {
	db, err := testutil.GetMySQLTestConn()
	require.NoError(t, err)
	defer db.Close()

	setupSQL := `
		CREATE TABLE users (
			id INT PRIMARY KEY,
			name VARCHAR(255),
			email VARCHAR(255),
			active BOOLEAN
		);

		CREATE TABLE posts (
			id INT PRIMARY KEY,
			title VARCHAR(255),
			content TEXT,
			created_at TIMESTAMP
		);

		CREATE VIEW active_users AS
		SELECT * FROM users WHERE active = 1;

		CREATE VIEW recent_posts AS
		SELECT * FROM posts WHERE created_at > NOW() - INTERVAL 7 DAY;
	`
	_, err = db.Exec(setupSQL)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, err := db.Exec(`
			DROP VIEW IF EXISTS active_users;
			DROP VIEW IF EXISTS recent_posts;
			DROP TABLE IF EXISTS users;
			DROP TABLE IF EXISTS posts;
		`)
		require.NoError(t, err)
	})

	my := New()

	s := &schema.Schema{}

	err = my.InspectViews(db, s)
	require.NoError(t, err)

	expectedViews := []*schema.View{
		{
			Name:       "active_users",
			Columns:    []string{"id", "name", "email", "active"},
			Definition: "select `users`.`id` AS `id`,`users`.`name` AS `name`,`users`.`email` AS `email`,`users`.`active` AS `active` from `users` where (`users`.`active` = 1)",
		},
		{
			Name:       "recent_posts",
			Columns:    []string{"id", "title", "content", "created_at"},
			Definition: "select `posts`.`id` AS `id`,`posts`.`title` AS `title`,`posts`.`content` AS `content`,`posts`.`created_at` AS `created_at` from `posts` where (`posts`.`created_at` > (now() - interval 7 day))",
		},
	}
	require.Equal(t, expectedViews, s.Views)
}
