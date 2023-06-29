package loader

import (
	"database/sql"
	"encoding/json"
	"log"
	"time"

	"github.com/CrocSwap/graphcache-go/tables"

	_ "github.com/mattn/go-sqlite3"
)

var db_string = "./artifacts/db/blank.db"

type Swap struct {
	Swap     string `json:"swap"`
	SwapTime  int `json:"swap_time"`
	SwapId 	string `json:"swap_id"`
	Id 			int `json:"id"`
}
// TODO: clean this up

func logError(msg string, err error) {
	if err != nil {

		log.Println("[DB]: ", msg, " Error: ", err)
	}
}

func GetLatestSwapTime() (int, error) {
	db, err := sql.Open("sqlite3", db_string)
	if err != nil {
		logError("Error connecting db ", err)
		return 0, err
	}
	defer db.Close()

	stmt, err := db.Prepare(`
		SELECT swap_time
		FROM swaps
		ORDER BY swap_time DESC
		LIMIT 1;
		`)

	if err != nil {
		logError("Error prepping query ", err)
		return 0, err
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err = rows.Err(); err != nil {
		return 0, err
	}

	var swapTime string
	for rows.Next() {
		err := rows.Scan(&swapTime)
		if err != nil {
			logError("Error scanning ", err)
			return 0, err
		}
	}
	if swapTime == "" {
		log.Println("[DB]: No swap time found, returning January 1, 2022 GMT")
		return 1640995200, nil

	}

	t, err := time.Parse(time.RFC3339, swapTime)
	if err != nil {
		logError("Error parsing time ", err)
		return 0, err
	}

	return int(t.Unix()), nil

}
func queryFromDB(startTime int, endTime int, isAsc bool) (*sql.Rows, error) {
	db, err := sql.Open("sqlite3", db_string)

	if err != nil {
		logError("Error connecting db ", err)
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


func SaveAggEventSubgraphArrayToDB(arr []tables.UniSwapSubGraph) { 
	db, err := sql.Open("sqlite3", db_string)
	if err != nil {
		logError("Error connecting db ", err)
		return
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		logError("Error beginning transaction ", err)
		return
	}

	stmt, err := tx.Prepare(`

		INSERT INTO swaps(swap, swap_time, swap_id)
		VALUES(?, ?, ?)
		`)
	if err != nil {
		logError("Error preparing statement ", err)
		return 
	}
	defer stmt.Close()

	// covert struct row to json

	for _, row := range arr {
		jsonData, err := json.Marshal(row)
		if err != nil {
			log.Fatal(err)
			return
		}
		_, err = stmt.Exec(jsonData, row.Timestamp, row.ID)
		if err != nil {
			logError("Error executing statement ", err)
			return
		}
	}

	tx.Commit()

	log.Println("[DB]: Uniswap Swaps saved to DB: ", len(arr))



}




