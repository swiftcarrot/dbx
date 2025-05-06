package testutil

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

const MySQLTestConnString = "root@(localhost:3306)/dbx_test?parseTime=true&multiStatements=true"

func GetMySQLTestConn() (*sql.DB, error) {
	db, err := sql.Open("mysql", MySQLTestConnString)
	if err != nil {
		log.Fatal(err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	return db, err
}
