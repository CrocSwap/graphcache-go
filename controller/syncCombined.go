package controller

import (
	"log"
	"time"

	"github.com/CrocSwap/graphcache-go/loader"
	"github.com/CrocSwap/graphcache-go/types"
)

type CombinedSubgraphSyncer struct {
	cntr       *ControllerOverNetwork
	cfg        loader.SyncChannelConfig
	channels   syncChannels
	lastBlocks loader.CombinedStartBlocks
}

func NewCombinedSubgraphSyncer(controller *Controller, chainConfig loader.ChainConfig, network types.NetworkName) *CombinedSubgraphSyncer {
	start := SubgraphStartBlocks{
		Bal:   0,
		Swaps: 0,
		Aggs:  0,
	}
	return NewCombinedSubgraphSyncerAtStart(controller, chainConfig, network, start)
}

func NewCombinedSubgraphSyncerAtStart(controller *Controller, chainConfig loader.ChainConfig, network types.NetworkName, startBlocks SubgraphStartBlocks) *CombinedSubgraphSyncer {
	sync := makeCombinedSubgraphSyncer(controller, chainConfig, network)
	sync.lastBlocks.Bal = startBlocks.Bal
	sync.lastBlocks.Swap = startBlocks.Swaps
	sync.lastBlocks.Agg = startBlocks.Aggs
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

const COMBINED_SUBGRAPH_POLL_SECS = 3

func (s *CombinedSubgraphSyncer) PollSubgraphUpdates() {
	s.syncLoop(false)
}

func (s *CombinedSubgraphSyncer) syncStart() {
	s.syncLoop(true)
	log.Printf("Startup subgraph sync done on chainId=%d", s.cntr.chainCfg.ChainID)
}

const MAX_BLOCK = 999999999

func (s *CombinedSubgraphSyncer) syncLoop(startupSync bool) {
	lastSyncBlock := 0
	for {
		newRows := false
		syncBlock, comboData, err := loader.CombinedQuery(s.cfg, s.lastBlocks, MAX_BLOCK)
		if err != nil {
			log.Println("Warning unable to send combined query:", err.Error())
			time.Sleep(COMBINED_SUBGRAPH_POLL_SECS * time.Second)
			continue
		}

		lastObsSwaps, hasMoreSwaps, errSwaps := s.channels.swaps.IngestEntries(comboData.Swaps, s.lastBlocks.Swap, syncBlock)
		lastObsAgg, hasMoreAggs, errAggs := s.channels.aggs.IngestEntries(comboData.Aggs, s.lastBlocks.Agg, syncBlock)
		lastObsBal, hasMoreBals, errBals := s.channels.bal.IngestEntries(comboData.Bals, s.lastBlocks.Bal, syncBlock)

		lastObsLiq, hasMoreLiqs, errLiqs := s.channels.liq.IngestEntries(comboData.Liqs, s.lastBlocks.Liq, syncBlock)
		lastObsKo, hasMoreKos, errKos := s.channels.ko.IngestEntries(comboData.Kos, s.lastBlocks.Ko, syncBlock)
		lastObsFees, hasMoreFees, errFees := s.channels.fees.IngestEntries(comboData.Fees, s.lastBlocks.Fee, syncBlock)

		if errSwaps != nil || errAggs != nil || errBals != nil || errLiqs != nil || errKos != nil || errFees != nil {
			log.Println("Warning unable to ingest entries:", errSwaps, errAggs, errBals, errLiqs, errKos, errFees)
			time.Sleep(COMBINED_SUBGRAPH_POLL_SECS * time.Second)
			continue
		}

		if lastObsSwaps > s.lastBlocks.Swap || lastObsAgg > s.lastBlocks.Agg || lastObsBal > s.lastBlocks.Bal || lastObsLiq > s.lastBlocks.Liq || lastObsKo > s.lastBlocks.Ko || lastObsFees > s.lastBlocks.Fee {
			newRows = true
		}
		if lastObsSwaps > s.lastBlocks.Swap {
			s.lastBlocks.Swap = lastObsSwaps
		}
		if lastObsAgg > s.lastBlocks.Agg {
			s.lastBlocks.Agg = lastObsAgg
		}
		if lastObsBal > s.lastBlocks.Bal {
			s.lastBlocks.Bal = lastObsBal
		}
		if lastObsLiq > s.lastBlocks.Liq {
			s.lastBlocks.Liq = lastObsLiq
		}
		if lastObsKo > s.lastBlocks.Ko {
			s.lastBlocks.Ko = lastObsKo
		}
		if lastObsFees > s.lastBlocks.Fee {
			s.lastBlocks.Fee = lastObsFees
		}

		if newRows {
			log.Printf("Sync step. Swap: %d Agg: %d Bal: %d Liq: %d Ko: %d Fee: %d", s.lastBlocks.Swap, s.lastBlocks.Agg, s.lastBlocks.Bal, s.lastBlocks.Liq, s.lastBlocks.Ko, s.lastBlocks.Fee)
		}

		// If no more data to backfill, either sleep or exit if it's a startup sync.
		if !hasMoreSwaps && !hasMoreAggs && !hasMoreBals && !hasMoreLiqs && !hasMoreKos && !hasMoreFees {
			if startupSync {
				break
			}
			if syncBlock > lastSyncBlock {
				log.Printf("New subgraph time %d", syncBlock)
				lastSyncBlock = syncBlock
			}
			time.Sleep(COMBINED_SUBGRAPH_POLL_SECS * time.Second)
		}
	}
}
