package loader

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)


type Swap struct {
	Swap     string `json:"swap"`
	SwapTime  int `json:"swap_time"`
	SwapId 	string `json:"swap_id"`
	Id 			int `json:"id"`
}

func queryFromDB(startTime int, endTime int, isAsc bool) (*sql.Rows, error) {
	db, err := sql.Open("sqlite3", "./artifacts/db/mydatabase.db")

	if err != nil {
		fmt.Println(`query error `, err)
		return nil, err
	}
	defer db.Close()

	order := "ASC"
	if !isAsc {
		order = "DESC"
	}

	stmt, err := db.Prepare(`
		SELECT *
		FROM swaps
		WHERE swap_time >= ? AND swap_time <= ?
		ORDER BY swap_time ` + order + `
		LIMIT 10000
		`)


	if err != nil {
		fmt.Println(`query error `, err)
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(startTime, endTime)
	if err != nil {
		return nil, err
	}


	if err = rows.Err(); err != nil {
		return nil, err
	}

	return rows, nil
}

