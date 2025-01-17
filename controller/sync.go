package controller

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/CrocSwap/graphcache-go/loader"
	"github.com/CrocSwap/graphcache-go/tables"
	"github.com/CrocSwap/graphcache-go/types"
)

type SubgraphSyncer interface {
	ChainId() types.ChainId
	PollSubgraphUpdates()
	SetStartBlocks(startBlocks loader.SubgraphStartBlocks)
	IngestEntries(table string, entriesData []byte, startBlock, endBlock int) (lastObs int, hasMore bool, err error)
}

type NormalSubgraphSyncer struct {
	cntr           *ControllerOverNetwork
	cfg            loader.SyncChannelConfig
	lastSyncBlock  int
	lookbackBlocks int
	channels       syncChannels
	startBlocks    loader.SubgraphStartBlocks
}

type syncChannels struct {
	bal   loader.SyncChannel[tables.Balance, tables.BalanceSubGraph]
	liq   loader.SyncChannel[tables.LiqChange, tables.LiqChangeSubGraph]
	swaps loader.SyncChannel[tables.Swap, tables.SwapSubGraph]
	fees  loader.SyncChannel[tables.FeeChange, tables.FeeChangeSubGraph]
	aggs  loader.SyncChannel[tables.AggEvent, tables.AggEventSubGraph]
}

func NewSubgraphSyncer(controller *Controller, chainConfig loader.ChainConfig, network types.NetworkName, startupCache string) *NormalSubgraphSyncer {
	start := loader.SubgraphStartBlocks{}
	return NewSubgraphSyncerAtStart(controller, chainConfig, network, start, startupCache)
}

func NewSubgraphSyncerAtStart(controller *Controller, chainConfig loader.ChainConfig, network types.NetworkName, startBlocks loader.SubgraphStartBlocks, startupCache string) *NormalSubgraphSyncer {
	sync := makeSubgraphSyncer(controller, chainConfig, network)
	sync.startBlocks = startBlocks
	if startupCache != "" {
		LoadStartupCache(startupCache, &sync)
	}
	syncNotif := make(chan bool, 1)
	go sync.syncStart(syncNotif)
	<-syncNotif
	return &sync
}

func makeSubgraphSyncer(controller *Controller, chainConfig loader.ChainConfig, network types.NetworkName) NormalSubgraphSyncer {
	cfg := loader.SyncChannelConfig{
		Chain:   chainConfig,
		Network: network,
	}
	netCntr := controller.OnNetwork(network)

	return NormalSubgraphSyncer{
		cntr:     netCntr,
		cfg:      cfg,
		channels: makeSyncChannels(netCntr, cfg),
	}
}

// Used because subgraph synchronization is not observed to be non-atomic
// between meta latest time and updating individual tables. Gives the subraph
// indexer time to index the incremental rows
const SUBGRAPH_SYNC_DELAY = 1

func (s *NormalSubgraphSyncer) PollSubgraphUpdates() {
	pollInterval := DEFAULT_SUBGRAPH_POLL_SECS
	if os.Getenv("SUBGRAPH_POLL_SECS") != "" {
		pollSecs, err := strconv.Atoi(os.Getenv("SUBGRAPH_POLL_SECS"))
		if err != nil {
			log.Panicln("Invalid SUBGRAPH_POLL_SECS value", os.Getenv("SUBGRAPH_POLL_SECS"))
		}
		pollInterval = time.Duration(pollSecs) * time.Second
	}
	for {
		time.Sleep(pollInterval)
		hasMore, _ := s.checkNewSubgraphSync()
		if hasMore {
			log.Printf("New subgraph block %d", s.lastSyncBlock)
		}
	}
}

func (s *NormalSubgraphSyncer) checkNewSubgraphSync() (bool, error) {
	metaBlock, err := loader.LatestSubgraphBlock(s.cfg)
	if err != nil {
		log.Println("Warning unable to sync subgraph meta query " + err.Error())
		return false, err
	}

	if metaBlock > s.lastSyncBlock {
		time.Sleep(SUBGRAPH_SYNC_DELAY * time.Second)
		s.syncStep(metaBlock)
		return true, nil
	}
	return false, nil
}

func (s *NormalSubgraphSyncer) syncStart(notif chan bool) {
	syncBlock, err := loader.LatestSubgraphBlock(s.cfg)

	if err != nil || syncBlock == 0 {
		log.Fatalf("Subgraph not responding from %s", s.cntr.chainCfg.Subgraph)
	}

	s.syncStep(syncBlock)
	log.Printf("Startup subgraph sync done on chainId=%d", s.cntr.chainCfg.ChainID)
	notif <- true
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
		swaps: syncSwap,
		fees:  syncFee,
		aggs:  syncAgg,
	}
}

func (s *NormalSubgraphSyncer) syncStep(syncBlock int) {
	// We use the second to last previous sync time. This makes sure that every time
	// window is sycn'd for a second time on the next block. This is necessary to prevent
	// table synchronization issues where a window isn't fully synced on a table during the
	// first pass on the block
	startBlock := s.lookbackBlocks + 1

	var wg sync.WaitGroup

	const N_CHANNELS = 6
	wg.Add(N_CHANNELS)

	go s.channels.swaps.SyncTableToSubgraphWG(maxBlock(startBlock, s.startBlocks.Swaps), syncBlock, &wg)
	go s.channels.aggs.SyncTableToSubgraphWG(maxBlock(startBlock, s.startBlocks.Aggs), syncBlock, &wg)
	go s.channels.bal.SyncTableToSubgraphWG(maxBlock(startBlock, s.startBlocks.Bal), syncBlock, &wg)

	go s.channels.liq.SyncTableToSubgraphWG(maxBlock(startBlock, s.startBlocks.Liq), syncBlock, &wg)
	go s.channels.fees.SyncTableToSubgraphWG(maxBlock(startBlock, s.startBlocks.Fee), syncBlock, &wg)

	wg.Wait()

	s.lookbackBlocks = s.lastSyncBlock
	s.lastSyncBlock = syncBlock
}

func (s *NormalSubgraphSyncer) SetStartBlocks(startBlocks loader.SubgraphStartBlocks) {
	s.startBlocks = startBlocks
}

func (s *NormalSubgraphSyncer) IngestEntries(table string, entriesData []byte, startBlock, endBlock int) (int, bool, error) {
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

func (s *NormalSubgraphSyncer) ChainId() types.ChainId {
	return types.IntToChainId(s.cfg.Chain.ChainID)
}

func maxBlock(startBlock int, laterBlock int) int {
	if startBlock > laterBlock {
		return startBlock
	}
	return laterBlock
}
