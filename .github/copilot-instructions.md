# Project coding standards

## General guidelines

- Use the Go module `github.com/swiftcarrot/dbx`
- Do not include comments in the code unless they are necessary to explain complex logic or non-obvious functionality
- Do not create variables used only once

## Tests writing guidelines

- Follow general guidelines
- Always use `github.com/stretchr/testify/require`
- Avoid custom error messages
- Avoid using `t.Run`, always write separated test functions
- Avoid using multiple equality checks for a struct's fields, use `require.Equal` to compare the entire struct directly, for example:
    ```go
	require.Equal(t, []*schema.View{
		{
			Schema:     "public",
			Name:       "view_test_orders_detail",
			Definition: def2,
			Columns:    []string{"order_id", "item_name", "quantity", "price", "total_price"},
		},
		{
			Schema:     "public",
			Name:       "view_test_simple",
			Definition: def1,
			Columns:    []string{"id", "name", "price"},
		},
	}, s.Views)
    ```
- Use `t.Cleanup` to clean up any database objects created during each test, for example:
	```go
	t.Cleanup(func() {
		_, err := db.Exec(`
			DROP VIEW IF EXISTS active_users;
			DROP VIEW IF EXISTS recent_posts;
			DROP TABLE IF EXISTS users;
			DROP TABLE IF EXISTS posts;
		`)
		require.NoError(t, err)
	})
	```
- Avoid using `defer db.Close()` in test. Explicitly closing the connection is unnecessary and may cause issues in test environments.
- Avoid writing redundant test cases for the same feature or behavior. Each test should validate a unique aspect of the functionality, such as distinct inputs, edge cases, or error conditions. Group related tests logically and use parameterized tests or test tables in Go to cover similar scenarios efficiently.
