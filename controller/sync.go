package controller

import (
	"log"
	"time"

	"github.com/CrocSwap/graphcache-go/loader"
	"github.com/CrocSwap/graphcache-go/tables"
	"github.com/CrocSwap/graphcache-go/types"
)

type SubgraphSyncer struct {
	cntr         *ControllerOverNetwork
	cfg          loader.SyncChannelConfig
	lastSyncTime int
	lookbackTime int
	channels     syncChannels
}
type syncChannels struct {
	bal   loader.SyncChannel[tables.Balance, tables.BalanceSubGraph]
	liq   loader.SyncChannel[tables.LiqChange, tables.LiqChangeSubGraph]
	ko    loader.SyncChannel[tables.KnockoutCross, tables.KnockoutCrossSubGraph]
	swaps loader.SyncChannel[tables.Swap, tables.SwapSubGraph]
	fees  loader.SyncChannel[tables.FeeChange, tables.FeeChangeSubGraph]
	aggs  loader.SyncChannel[tables.AggEvent, tables.AggEventSubGraph]
}

func NewSubgraphSyncer(controller *Controller, chainConfig loader.ChainConfig, network types.NetworkName) SubgraphSyncer {
	sync := makeSubgraphSyncer(controller, chainConfig, network)
	syncNotif := make(chan bool, 1)
	go sync.syncStart(syncNotif)
	<-syncNotif
	return sync
}

func makeSubgraphSyncer(controller *Controller, chainConfig loader.ChainConfig, network types.NetworkName) SubgraphSyncer {
	cfg := loader.SyncChannelConfig{
		Chain:   chainConfig,
		Network: network,
	}
	netCntr := controller.OnNetwork(network)

	return SubgraphSyncer{
		cntr:     netCntr,
		cfg:      cfg,
		channels: makeSyncChannels(netCntr, cfg),
	}
}

const SUBGRAPH_POLL_SECS = 1

// Used because subgraph synchronization is not observed to be non-atomic
// between meta latest time and updating individual tables. Gives the subraph
// indexer time to index the incremental rows
const SUBGRAPH_SYNC_DELAY = 1

func (s *SubgraphSyncer) pollSubgraphUpdates() {
	for true {
		time.Sleep(SUBGRAPH_POLL_SECS * time.Second)
		hasMore, _ := s.checkNewSubgraphSync()
		if hasMore {
			log.Printf("New subgraph time %d", s.lastSyncTime)
		}
	}
}

func (s *SubgraphSyncer) checkNewSubgraphSync() (bool, error) {
	metaTime, err := loader.LatestSubgraphTime(s.cfg)
	if err != nil {
		log.Println("Warning unable to sync subgraph meta query " + err.Error())
		return false, err
	}

	if metaTime > s.lastSyncTime {
		time.Sleep(SUBGRAPH_SYNC_DELAY * time.Second)
		s.syncStep(metaTime)
		return true, nil
	}
	return false, nil
}

func (s *SubgraphSyncer) syncStart(notif chan bool) {
	syncTime, err := loader.LatestSubgraphTime(s.cfg)

	if err != nil {
		log.Fatalf("Subgraph not responding from %s", s.cntr.chainCfg.Subgraph)
	}

	s.syncStep(syncTime)
	log.Printf("Startup subgraph sync done on chainId=%d", s.cntr.chainCfg.ChainID)
	notif <- true

	s.pollSubgraphUpdates()
}

func (s *SubgraphSyncer) logSyncCycle(table string, nRows int) {
	if nRows > 0 {
		log.Printf("Sync %s subgraph on chainId=%d with rows=%d", table, s.cntr.chainCfg.ChainID, nRows)
	}
}

func makeSyncChannels(cntr *ControllerOverNetwork, cfg loader.SyncChannelConfig) syncChannels {
	cfg.Query = "./artifacts/graphQueries/liqchanges.query"
	tblLiq := tables.LiqChangeTable{}
	syncLiq := loader.NewSyncChannel[tables.LiqChange, tables.LiqChangeSubGraph](
		tblLiq, cfg, cntr.IngestLiqChange)

	cfg.Query = "./artifacts/graphQueries/swaps.query"
	tblSwap := tables.SwapsTable{}
	syncSwap := loader.NewSyncChannel[tables.Swap, tables.SwapSubGraph](
		tblSwap, cfg, cntr.IngestSwap)

	cfg.Query = "./artifacts/graphQueries/knockoutcrosses.query"
	tblKo := tables.KnockoutTable{}
	syncKo := loader.NewSyncChannel[tables.KnockoutCross, tables.KnockoutCrossSubGraph](
		tblKo, cfg, cntr.IngestKnockout)

	cfg.Query = "./artifacts/graphQueries/feechanges.query"
	tblFee := tables.FeeTable{}
	syncFee := loader.NewSyncChannel[tables.FeeChange, tables.FeeChangeSubGraph](
		tblFee, cfg, cntr.IngestFee)

	cfg.Query = "./artifacts/graphQueries/aggevent.query"
	tblAgg := tables.AggEventsTable{}
	syncAgg := loader.NewSyncChannel[tables.AggEvent, tables.AggEventSubGraph](
		tblAgg, cfg, cntr.IngestAggEvent)

	cfg.Query = "./artifacts/graphQueries/balances.query"
	tblBal := tables.BalanceTable{}
	syncBal := loader.NewSyncChannel[tables.Balance, tables.BalanceSubGraph](
		tblBal, cfg, cntr.IngestBalance)

	return syncChannels{
		bal:   syncBal,
		liq:   syncLiq,
		ko:    syncKo,
		swaps: syncSwap,
		fees:  syncFee,
		aggs:  syncAgg,
	}
}

func (s *SubgraphSyncer) syncStep(syncTime int) {
	// We use the second to last previous sync time. This makes sure that every time
	// window is sycn'd for a second time on the next block. This is necessary to prevent
	// table synchronization issues where a window isn't fully synced on a table during the
	// first pass on the block
	startTime := s.lookbackTime + 1

	nRows, _ := s.channels.swaps.SyncTableToSubgraph(startTime, syncTime)
	s.logSyncCycle("Swaps", nRows)

	nRows, _ = s.channels.liq.SyncTableToSubgraph(startTime, syncTime)
	s.logSyncCycle("LiqChanges", nRows)

	nRows, _ = s.channels.ko.SyncTableToSubgraph(startTime, syncTime)
	s.logSyncCycle("Knockout crosses", nRows)

	nRows, _ = s.channels.bal.SyncTableToSubgraph(startTime, syncTime)
	s.logSyncCycle("User Balances", nRows)

	nRows, _ = s.channels.fees.SyncTableToSubgraph(startTime, syncTime)
	s.logSyncCycle("Fee Changes", nRows)

	nRows, _ = s.channels.aggs.SyncTableToSubgraph(startTime, syncTime)
	s.logSyncCycle("Poll Agg Events", nRows)

	s.cntr.FlushSyncCycle(syncTime)
	s.lookbackTime = s.lastSyncTime
	s.lastSyncTime = syncTime
}
