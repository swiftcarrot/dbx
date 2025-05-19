package testutil

import (
	"database/sql"

	_ "github.com/lib/pq"
)

const PGTestConnString = "postgres://postgres:postgres@localhost:5432/dbx_test?sslmode=disable"

func GetPGTestConn() (*sql.DB, error) {
	db, err := sql.Open("postgres", PGTestConnString)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
