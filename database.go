package main

import (
	"database/sql"
	"log"
)

func openSqliteDb(path string) *sql.DB {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		log.Fatal(err)
	}
	return db
}
