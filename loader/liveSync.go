package loader

import (
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
	config           SyncChannelConfig
	tbl              tables.ITable[R, S]
}

type SyncChannelConfig struct {
	Chain   ChainConfig
	Network types.NetworkName
	Query   string
}

func NewSyncChannel[R any, S any](tbl tables.ITable[R, S], config SyncChannelConfig,
	consumeFn func(R)) SyncChannel[R, S] {
	return SyncChannel[R, S]{
		LastObserved:     0,
		EarliestObserved: 1000 * 1000 * 1000 * 1000,
		idsObserved:      make(map[string]bool),
		consumeFn:        consumeFn,
		config:           config,
		tbl:              tbl,
	}
}

func (s *SyncChannel[R, S]) SyncTableFromDb(dbPath string) {
	db := openSqliteDb(dbPath)
	query := fmt.Sprintf("SELECT * FROM %s WHERE network == '%s' ORDER BY time ASC",
		s.tbl.SqlTableName(),
		string(s.config.Network))

	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		entry := s.tbl.ReadSqlRow(rows)
		s.ingestEntry(entry)
	}

	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
}

func (s *SyncChannel[R, S]) SyncTableToSubgraph(isAsc bool) (int, error) {
	query := readQueryPath(s.config.Query)

	startTime := 0
	if !isAsc {
		startTime = int(time.Now().Unix())
	}

	hasMore := true
	prevObs := startTime
	nIngested := 0

	log.Printf("Starting subgraph query %s", s.config.Query)
	for hasMore {
		resp, err := queryFromSubgraph(s.config.Chain, query, prevObs, isAsc)
		if err != nil {
			return nIngested, err
		}

		entries, err := s.tbl.ParseSubGraphResp(resp)
		if err != nil {
			log.Println("Warning subgraph request decode error " + err.Error())
			return nIngested, err
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

		log.Printf("Loaded %d rows from subgraph from query %s up to time=%d",
			nIngested, s.config.Query, prevObs)
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
