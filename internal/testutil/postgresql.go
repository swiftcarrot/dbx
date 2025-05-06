package testutil

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

const PGTestConnString = "postgres://postgres:postgres@localhost:5432/dbx_test?sslmode=disable"

func GetPGTestConn() (*sql.DB, error) {
	db, err := sql.Open("postgres", PGTestConnString)
	if err != nil {
		log.Fatal(err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	return db, err
}
