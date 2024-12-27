package controller

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/CrocSwap/graphcache-go/loader"
	"github.com/CrocSwap/graphcache-go/types"
)

type CombinedSubgraphSyncer struct {
	cntr       *ControllerOverNetwork
	cfg        loader.SyncChannelConfig
	channels   syncChannels
	lastBlocks loader.SubgraphStartBlocks
}

func NewCombinedSubgraphSyncer(controller *Controller, chainConfig loader.ChainConfig, network types.NetworkName, startupCacheDir string, startupCache string) *CombinedSubgraphSyncer {
	start := loader.SubgraphStartBlocks{}
	return NewCombinedSubgraphSyncerAtStart(controller, chainConfig, network, start, startupCache)
}

func NewCombinedSubgraphSyncerAtStart(controller *Controller, chainConfig loader.ChainConfig, network types.NetworkName, startBlocks loader.SubgraphStartBlocks, startupCache string) *CombinedSubgraphSyncer {
	sync := makeCombinedSubgraphSyncer(controller, chainConfig, network)
	sync.lastBlocks.Bal = startBlocks.Bal
	sync.lastBlocks.Swaps = startBlocks.Swaps
	sync.lastBlocks.Aggs = startBlocks.Aggs
	sync.lastBlocks.Liq = startBlocks.Liq
	if startupCache != "" {
		LoadStartupCache(startupCache, &sync)
	}
	sync.syncStart()
	return &sync
}

func makeCombinedSubgraphSyncer(controller *Controller, chainConfig loader.ChainConfig, network types.NetworkName) CombinedSubgraphSyncer {
	cfg := loader.SyncChannelConfig{
		Chain:   chainConfig,
		Network: network,
	}
	netCntr := controller.OnNetwork(network)

	return CombinedSubgraphSyncer{
		cntr:     netCntr,
		cfg:      cfg,
		channels: makeSyncChannels(netCntr, cfg),
	}
}

const DEFAULT_SUBGRAPH_POLL_SECS = 4 * time.Second

func (s *CombinedSubgraphSyncer) PollSubgraphUpdates() {
	s.syncLoop(false)
}

func (s *CombinedSubgraphSyncer) syncStart() {
	s.syncLoop(true)
	log.Printf("Startup subgraph sync done on chainId=%d", s.cntr.chainCfg.ChainID)
}

const MAX_BLOCK = 999999999

func (s *CombinedSubgraphSyncer) syncLoop(startupSync bool) {
	pollInterval := DEFAULT_SUBGRAPH_POLL_SECS
	if os.Getenv("SUBGRAPH_POLL_SECS") != "" {
		pollSecs, err := strconv.Atoi(os.Getenv("SUBGRAPH_POLL_SECS"))
		if err != nil {
			log.Panicln("Invalid SUBGRAPH_POLL_SECS value", os.Getenv("SUBGRAPH_POLL_SECS"))
		}
		pollInterval = time.Duration(pollSecs) * time.Second
	}

	lastSyncBlock := 0
	for {
		newRows := false
		syncBlock, comboData, err := loader.CombinedQuery(s.cfg, s.lastBlocks, MAX_BLOCK)
		if err != nil {
			log.Println("Warning unable to send combined query:", err.Error())
			time.Sleep(pollInterval)
			continue
		}

		lastObsSwaps, hasMoreSwaps, errSwaps := s.channels.swaps.IngestEntries(comboData.Swaps, s.lastBlocks.Swaps, syncBlock)
		lastObsAgg, hasMoreAggs, errAggs := s.channels.aggs.IngestEntries(comboData.Aggs, s.lastBlocks.Aggs, syncBlock)
		lastObsBal, hasMoreBals, errBals := s.channels.bal.IngestEntries(comboData.Bals, s.lastBlocks.Bal, syncBlock)

		lastObsLiq, hasMoreLiqs, errLiqs := s.channels.liq.IngestEntries(comboData.Liqs, s.lastBlocks.Liq, syncBlock)
		lastObsFees, hasMoreFees, errFees := s.channels.fees.IngestEntries(comboData.Fees, s.lastBlocks.Fee, syncBlock)

		if errSwaps != nil || errAggs != nil || errBals != nil || errLiqs != nil || errFees != nil {
			log.Println("Warning unable to ingest entries:", errSwaps, errAggs, errBals, errLiqs, errFees)
			time.Sleep(pollInterval)
			continue
		}

		if lastObsSwaps > s.lastBlocks.Swaps || lastObsAgg > s.lastBlocks.Aggs || lastObsBal > s.lastBlocks.Bal || lastObsLiq > s.lastBlocks.Liq || lastObsFees > s.lastBlocks.Fee {
			newRows = true
		}
		// Abnormal case, should sleep it off. But also if this happens during non-startup sync, not sleeping
		// would repeat the query immediately (because `hasMore` needs to be true) which might lead to spam.
		allowEmpty := os.Getenv("ALLOW_EMPTY_SUBGRAPH_TABLES") == "true"
		if (lastObsSwaps == 0 || lastObsAgg == 0 || lastObsBal == 0 || lastObsLiq == 0 || lastObsFees == 0) && !allowEmpty {
			log.Println("Warning: subgraph returned no rows for one or more tables")
			time.Sleep(pollInterval)
		}

		if lastObsSwaps > s.lastBlocks.Swaps {
			s.lastBlocks.Swaps = lastObsSwaps
		}
		if lastObsAgg > s.lastBlocks.Aggs {
			s.lastBlocks.Aggs = lastObsAgg
		}
		if lastObsBal > s.lastBlocks.Bal {
			s.lastBlocks.Bal = lastObsBal
		}
		if lastObsLiq > s.lastBlocks.Liq {
			s.lastBlocks.Liq = lastObsLiq
		}
		if lastObsFees > s.lastBlocks.Fee {
			s.lastBlocks.Fee = lastObsFees
		}

		if newRows {
			log.Printf("Sync step. Swap: %d Agg: %d Bal: %d Liq: %d Fee: %d", s.lastBlocks.Swaps, s.lastBlocks.Aggs, s.lastBlocks.Bal, s.lastBlocks.Liq, s.lastBlocks.Fee)
		}

		// If no more data to backfill, either sleep or exit if it's a startup sync.
		if !hasMoreSwaps && !hasMoreAggs && !hasMoreBals && !hasMoreLiqs && !hasMoreFees {
			if startupSync {
				break
			}
			if syncBlock > lastSyncBlock {
				log.Printf("New subgraph block %d", syncBlock)
				lastSyncBlock = syncBlock
			}
			time.Sleep(pollInterval)
		}
	}
}

func (s *CombinedSubgraphSyncer) SetStartBlocks(startBlocks loader.SubgraphStartBlocks) {
	s.lastBlocks = startBlocks
}

func (s *CombinedSubgraphSyncer) IngestEntries(table string, entriesData []byte, startBlock, endBlock int) (int, bool, error) {
	switch table {
	case "swaps":
		return s.channels.swaps.IngestEntries(entriesData, startBlock, endBlock)
	case "aggEvents":
		return s.channels.aggs.IngestEntries(entriesData, startBlock, endBlock)
	case "liquidityChanges":
		return s.channels.liq.IngestEntries(entriesData, startBlock, endBlock)
	case "feeChanges":
		return s.channels.fees.IngestEntries(entriesData, startBlock, endBlock)
	case "userBalances":
		return s.channels.bal.IngestEntries(entriesData, startBlock, endBlock)
	default:
		log.Fatal("Warning: unknown table name in subgraph ingest", table)
	}
	return 0, false, fmt.Errorf("unknown table name in subgraph ingest")
}

func (s *CombinedSubgraphSyncer) ChainId() types.ChainId {
	return types.IntToChainId(s.cfg.Chain.ChainID)
}
