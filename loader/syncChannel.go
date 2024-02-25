package loader

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/CrocSwap/graphcache-go/tables"
	"github.com/CrocSwap/graphcache-go/types"
)

type SyncChannel[R any, S any] struct {
	RowsIngested int
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
		idsObserved: make(map[string]bool),
		consumeFn:   consumeFn,
		config:      config,
		tbl:         tbl,
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

	if result.Block.Time == 0 {
		log.Println("Warning subgraph latest block time is 0. Retrying")
		return LatestSubgraphTime(cfg)
	} else {
		return result.Block.Time, nil
	}
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

func (s *SyncChannel[R, S]) SyncTableToSubgraphWG(startTime int, endTime int, wg *sync.WaitGroup) (int, error) {
	defer wg.Done()
	return s.SyncTableToSubgraph(startTime, endTime)
}

func (s *SyncChannel[R, S]) SyncTableToSubgraph(startTime int, endTime int) (int, error) {
	query := readQueryPath(s.config.Query)

	lastObs := startTime

	hasMore := true
	nIngested := 0

	for hasMore {
		hasMore = false

		queryStartTime := lastObs
		queryEndTime := endTime

		resp, err := queryFromSubgraph(s.config.Chain, query, queryStartTime, queryEndTime, true)

		if err != nil {
			log.Println("Warning subgraph request error " + err.Error())
			return nIngested, err
		}

		entries, err := s.tbl.ParseSubGraphResp(resp)
		if err != nil {
			log.Println("Warning subgraph request decode error " + err.Error())
			return nIngested, err
		}

		for _, entry := range entries {
			row := s.tbl.ConvertSubGraphRow(entry, string(s.config.Network))
			isFreshPoint, eventTime := s.ingestEntry(row)

			if isFreshPoint {
				nIngested += 1
				hasMore = true

				if eventTime > lastObs {
					lastObs = eventTime
				}
			}
		}

		if nIngested > 0 {
			log.Printf("Loaded %d rows from subgraph from query %s on time=%d-%d",
				nIngested, s.config.Query, queryStartTime, queryEndTime)
		}
	}
	return nIngested, nil
}

func (s *SyncChannel[R, S]) ingestEntry(r R) (bool, int) {
	_, hasEntry := s.idsObserved[s.tbl.GetID(r)]

	if !hasEntry {
		s.idsObserved[s.tbl.GetID(r)] = true
		s.consumeFn(r)
		return true, s.tbl.GetTime(r)
	} else {
		return false, -1
	}
}
