package loader

import (
	"encoding/json"
	"fmt"

	"github.com/CrocSwap/graphcache-go/tables"
	"github.com/CrocSwap/graphcache-go/types"
)

type SyncChannel struct {
	lastObserved int
	idsObserved  map[string]bool
	consumeFn    func(tables.Balance)
	config       SyncChannelConfig
}

type SyncChannelConfig struct {
	Chain   ChainConfig
	Network types.NetworkName
	Query   string
}

func (s *SyncChannel) ingestEntry(b tables.Balance) {
	if b.Time > s.lastObserved {
		s.lastObserved = b.Time
	}

	_, hasEntry := s.idsObserved[b.ID]
	if !hasEntry {
		s.idsObserved[b.ID] = true
		s.consumeFn(b)
	}
}

func NewSyncChannel(config SyncChannelConfig, consumeFn func(tables.Balance)) SyncChannel {
	return SyncChannel{
		lastObserved: 0,
		idsObserved:  make(map[string]bool),
		consumeFn:    consumeFn,
		config:       config,
	}
}

func (s *SyncChannel) SyncTableFromDb(dbPath string) {
	db := openSqliteDb(dbPath)
	tables.LoadTokenBalancesSql(db, string(s.config.Network), s.ingestEntry)
}

func (s *SyncChannel) SyncTableToSubgraph() error {
	var parsed tables.BalanceSubGraphResp
	query := readQueryPath(s.config.Query)

	hasMore := true
	numObs := len(s.idsObserved)

	for hasMore {
		resp, err := queryFromSubgraph(s.config.Chain, query, s.lastObserved)
		if err != nil {
			return err
		}

		err = json.Unmarshal(resp, &parsed)
		if err != nil {
			fmt.Println("Warning subgraph request decode error " + err.Error())
			return err
		}

		for _, entry := range parsed.Data.UserBalances {
			row := tables.ConvertBalanceToSql(entry, string(s.config.Network))
			s.ingestEntry(row)
		}

		hasMore = len(parsed.Data.UserBalances) > numObs
		numObs = len(parsed.Data.UserBalances)
	}
	return nil
}
