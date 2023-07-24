package loader

import (
	"database/sql"
	"log"
)

func logError(msg string, err error) {
	if err != nil {

		log.Println("[DB]: ", msg, " Error: ", err)
	}
}

func QueryFromDB(startTime int, endTime int, isAsc bool, dbString string) (*sql.Rows, error) {
	db_, err := sql.Open("sqlite3", dbString)

	if err != nil {
		logError("Error connecting db ", err)
		return nil, err
	}
	defer db_.Close()

	order := "ASC"
	if !isAsc {
		order = "DESC"
	}

	stmt, err := db_.Prepare(`
		SELECT *
		FROM swaps
		WHERE swap_time >= ? AND swap_time <= ?
		ORDER BY swap_time ` + order + `
		LIMIT 10000
		`)


	if err != nil {
		logError("Error prepping query ", err)
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(startTime, endTime)
	if err != nil {
		logError("Error querying db ", err)
		return nil, err
	}


	if err = rows.Err(); err != nil {
		logError("Error iterating rows ", err)
		return nil, err
	}

	return rows, nil
}