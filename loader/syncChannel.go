package loader

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/CrocSwap/graphcache-go/tables"
	"github.com/CrocSwap/graphcache-go/types"
)

type SyncChannel[R any, S any] struct {
	LastObserved     int
	EarliestObserved int
	RowsIngested     int
	idsObserved      map[string]bool
	consumeFn        func(R)
	saveToDBFn       func([]S)
	config           SyncChannelConfig
	tbl              tables.ITable[R, S]
}

type SyncChannelConfig struct {
	Chain   ChainConfig
	Network types.NetworkName
	Query   string
}

func NewSyncChannel[R any, S any](tbl tables.ITable[R, S], config SyncChannelConfig,
	consumeFn func(R), saveToDBFn func([]S)) SyncChannel[R, S] {
	return SyncChannel[R, S]{
		LastObserved:     0,
		EarliestObserved: 1000 * 1000 * 1000 * 1000,
		idsObserved:      make(map[string]bool),
		consumeFn:        consumeFn,
		saveToDBFn:       saveToDBFn,
		config:           config,
		tbl:              tbl,
	}
}

func LatestSubgraphTime(cfg SyncChannelConfig) (int, error) {
	cfg.Query = "./artifacts/graphQueries/meta.query"
	metaQuery := readQueryPath(cfg.Query)

	resp, err := queryFromSubgraph(cfg.Chain, metaQuery, 0, 0, false)
	if err != nil {
		return 0, err
	}

	result, err := parseSubGraphMeta(resp)
	if err != nil {
		return 0, err
	}
	return result.Block.Time, nil
}

type metaEntry struct {
	Block struct {
		Time   int    `json:"timestamp"`
		Number int    `json:"number"`
		Hash   string `json:"hash"`
	} `json:"block"`
}

type metaWrapper struct {
	Entry metaEntry `json:"_meta"`
}

type metaData struct {
	Data metaWrapper `json:"data"`
}

func parseSubGraphMeta(body []byte) (*metaEntry, error) {
	var parsed metaData

	err := json.Unmarshal(body, &parsed)
	if err != nil {
		return nil, err
	}
	return &parsed.Data.Entry, nil
}

func (s *SyncChannel[R, S]) SyncTableToDB(isAsc bool, startTime int, endTime int) (int, error) {

	prevObs := startTime
	if !isAsc {
		prevObs = endTime
	}

	hasMore := true
	nIngested := 0

	for hasMore {
		var db_resp *sql.Rows
		var err error
		if isAsc {
			db_resp, err = queryFromDB(prevObs, endTime, isAsc)
		} else {
			db_resp, err = queryFromDB(startTime, prevObs, isAsc)
		}

		// Iterate over the rows and process the data
		for db_resp.Next() {
			var id int
			var swap_string string
			var swap_id string
			var swap_time string
			err = db_resp.Scan(&id, &swap_string, &swap_time, &swap_id)
			if err != nil {
				fmt.Println(err)
			}

			var swap S
			err := json.Unmarshal([]byte(swap_string), &swap)
			if err != nil {
				fmt.Println("Error parsing JSON:", err)

			}
			row := s.tbl.ConvertSubGraphRow(swap, string(s.config.Network))
			s.ingestEntry(row)
			nIngested += 1
		}

		if isAsc {
			hasMore = s.LastObserved > prevObs
			prevObs = s.LastObserved
		} else {
			hasMore = s.EarliestObserved < prevObs
			prevObs = s.EarliestObserved
		}

		if nIngested > 0 {
			log.Printf("[Historical Syncer]: Loaded %d rows from subgraph from query %s up to time=%d - %s",
				nIngested, s.config.Query, prevObs, time.Unix(int64(prevObs), 0).String())
		}
	}
	return nIngested, nil
}

func (s *SyncChannel[R, S]) SyncTableToSubgraph(isAsc bool, startTime int, endTime int) (int, error) {
	query := readQueryPath(s.config.Query)

	prevObs := startTime
	if !isAsc {
		prevObs = endTime
	}

	hasMore := true
	nIngested := 0

	for hasMore {
		var resp []byte
		var err error

		if isAsc {
			resp, err = queryFromSubgraph(s.config.Chain, query, prevObs, endTime, isAsc)
		} else {
			resp, err = queryFromSubgraph(s.config.Chain, query, startTime, prevObs, isAsc)
		}

		if err != nil {
			return nIngested, err
		}

		entries, err := s.tbl.ParseSubGraphResp(resp)

		if err != nil {
			log.Println("Warning subgraph request decode error " + err.Error())
			return nIngested, err
		}

		if s.saveToDBFn != nil {
			s.saveToDBFn(entries)
		}

		for _, entry := range entries {
			row := s.tbl.ConvertSubGraphRow(entry, string(s.config.Network))
			s.ingestEntry(row)
			nIngested += 1
		}

		if isAsc {
			hasMore = s.LastObserved > prevObs
			prevObs = s.LastObserved
		} else {
			hasMore = s.EarliestObserved < prevObs
			prevObs = s.EarliestObserved
		}

		var logString string
		if s.saveToDBFn != nil {
			logString = "[Historical Syncer]:"
		} else {
			logString = "[Polling Syncer]:"
		}
		if nIngested > 0 {
			log.Printf("%s Loaded %d rows from subgraph from query %s up to time=%d - %s", logString,
				nIngested, s.config.Query, prevObs, time.Unix(int64(prevObs), 0).String())
		}
	}
	return nIngested, nil
}

func (s *SyncChannel[R, S]) ingestEntry(r R) {
	if s.tbl.GetTime(r) > s.LastObserved {
		s.LastObserved = s.tbl.GetTime(r)
	}
	if s.tbl.GetTime(r) < s.EarliestObserved {
		s.EarliestObserved = s.tbl.GetTime(r)
	}

	_, hasEntry := s.idsObserved[s.tbl.GetID(r)]
	if !hasEntry {
		s.idsObserved[s.tbl.GetID(r)] = true
		s.consumeFn(r)
	}
}
