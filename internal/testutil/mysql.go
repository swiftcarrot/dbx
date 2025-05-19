package testutil

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

const MySQLTestConnString = "root@(localhost:3306)/dbx_test?parseTime=true&multiStatements=true"

func GetMySQLTestConn() (*sql.DB, error) {
	db, err := sql.Open("mysql", MySQLTestConnString)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
