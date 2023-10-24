package db

import (
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/CrocSwap/graphcache-go/loader"
	"github.com/CrocSwap/graphcache-go/tables"
	_ "github.com/mattn/go-sqlite3"
)





func FetchUniswapAndSaveToShard(chainCfg loader.ChainConfig, shardPath string, startTime int, endTime int, source string) {
	trimmedFileName := strings.TrimSuffix(shardPath, ".db")
	shardFile := trimmedFileName + ".db"
	shardTempFile := trimmedFileName + "_temp.db"
	if(FileExistsInDir(shardFile)){
		log.Printf("[%s]: Shard already exists, skipping %s ", source, shardPath)
		return 
	}

	db_, err := sql.Open("sqlite3", shardTempFile) // Replace with the name of your database file
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	defer db_.Close()

	// Create the swaps table if it doesn't exist
	_, err = db_.Exec(`
		CREATE TABLE IF NOT EXISTS swaps (
			id INTEGER PRIMARY KEY,
			swap JSON,
			swap_time DATETIME,
			swap_id STRING UNIQUE
		);
	`)
	if err != nil {
		log.Fatalf("Failed to create swaps table: %v", err)
	}

	uniswapsTable := tables.UniSwapsTable{}
	total_swaps := 0

	for startTime < endTime {

		query := loader.ReadQueryPath("./artifacts/graphQueries/swaps.uniswap.query")
		response, err := loader.QueryFromSubgraph(chainCfg, query, startTime, endTime, true)

		if err != nil {
			log.Fatalf("Failed to send API request: %v", err)
			return 
		} 
	
		log.Printf("[%s]: Fetched swaps between %s and %s\n",source, time.Unix(int64(startTime), 0).Format("01/02/2006 15:04:05"), time.Unix(int64(endTime), 0).Format("01/02/2006 15:04:05"))
		// Check the status of the request
		if err == nil {
			swaps, ok := uniswapsTable.ParseSubGraphResp(response)
			if ok != nil {
				log.Fatal("Invalid response format")
				return
			}
			if(len(swaps) == 0){
				break
			}
			total_swaps += len(swaps)
			log.Printf("[%s]: Total swaps: %d", source, total_swaps)

			// Update the timestamp to be the timestamp of the last swap in the result
			lastSwap := swaps[len(swaps)-1]
			newStartTime, err := strconv.Atoi(lastSwap.Timestamp)
			startTime = newStartTime + 1

			tx, err := db_.Begin()
			if err != nil {
				log.Fatalf("Failed to start transaction: %v", err)
			}

			// Insert each swap into the database
			for _, swap := range swaps {
				swapData, err := json.Marshal(swap)
				if err != nil {
					log.Printf("Failed to marshal swap JSON: %v", err)
					continue
				}

				_, err = tx.Exec(`
					INSERT INTO swaps (swap, swap_time, swap_id)
					VALUES (?, ?, ?)
				`, string(swapData), swap.Timestamp, swap.ID)

				if err != nil {
					log.Printf("Failed to insert swap: %v", err)
					continue
				}
			}


			err = tx.Commit()
			if err != nil {
				log.Fatalf("Failed to commit transaction: %v", err)
			}

			// If there are no more swaps, break the loop
			if len(swaps) == 0 {
				break
			}
		} else {
			log.Printf("Query failed with status code %d", err)
			break
		}
	}

	var rename_error = os.Rename(shardTempFile, shardFile)
	if rename_error != nil {
		log.Fatalf("Failed to rename file: %v", err)
	}
	go UploadShardToBucket(shardFile)

}
