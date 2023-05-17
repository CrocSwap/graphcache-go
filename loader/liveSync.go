package loader

import (
	"fmt"
	"log"

	"github.com/CrocSwap/graphcache-go/tables"
	"github.com/CrocSwap/graphcache-go/types"
)

type SyncChannel[R any, S any] struct {
	lastObserved int
	idsObserved  map[string]bool
	consumeFn    func(R)
	config       SyncChannelConfig
	tbl          tables.ITable[R, S]
}

type SyncChannelConfig struct {
	Chain   ChainConfig
	Network types.NetworkName
	Query   string
}

func NewSyncChannel[R any, S any](tbl tables.ITable[R, S], config SyncChannelConfig,
	consumeFn func(R)) SyncChannel[R, S] {
	return SyncChannel[R, S]{
		lastObserved: 0,
		idsObserved:  make(map[string]bool),
		consumeFn:    consumeFn,
		config:       config,
		tbl:          tbl,
	}
}

func (s *SyncChannel[R, S]) SyncTableFromDb(dbPath string) {
	db := openSqliteDb(dbPath)
	query := fmt.Sprintf("SELECT * FROM %s WHERE network == '%s'",
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

func (s *SyncChannel[R, S]) SyncTableToSubgraph() error {
	var parsed tables.BalanceSubGraphResp
	query := readQueryPath(s.config.Query)

	hasMore := true
	numObs := len(s.idsObserved)

	for hasMore {
		resp, err := queryFromSubgraph(s.config.Chain, query, s.lastObserved)
		if err != nil {
			return err
		}

		entries, err := s.tbl.ParseSubGraphResp(resp)
		if err != nil {
			log.Println("Warning subgraph request decode error " + err.Error())
			return err
		}

		for _, entry := range entries {
			row := s.tbl.ConvertSubGraphRow(entry, string(s.config.Network))
			s.ingestEntry(row)
		}

		hasMore = len(parsed.Data.UserBalances) > numObs
		numObs = len(parsed.Data.UserBalances)
	}
	return nil
}

func (s *SyncChannel[R, S]) ingestEntry(r R) {
	if s.tbl.GetTime(r) > s.lastObserved {
		s.lastObserved = s.tbl.GetTime(r)
	}

	_, hasEntry := s.idsObserved[s.tbl.GetID(r)]
	if !hasEntry {
		s.idsObserved[s.tbl.GetID(r)] = true
		s.consumeFn(r)
	}
}
