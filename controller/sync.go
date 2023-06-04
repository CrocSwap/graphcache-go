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
}

func NewSubgraphSyncer(controller *Controller, chainConfig loader.ChainConfig, network types.NetworkName) SubgraphSyncer {
	sync := makeSubgraphSyncer(controller, chainConfig, network)
	go sync.syncStart()
	return sync
}

func NewSubgraphPriceSyncer(controller *Controller, chainConfig loader.ChainConfig, network types.NetworkName) SubgraphSyncer {
	sync := makeSubgraphSyncer(controller, chainConfig, network)
	go sync.syncPricingSwaps()
	return sync
}

func makeSubgraphSyncer(controller *Controller, chainConfig loader.ChainConfig, network types.NetworkName) SubgraphSyncer {
	cfg := loader.SyncChannelConfig{
		Chain:   chainConfig,
		Network: network,
	}
	netCntr := controller.OnNetwork(network)

	return SubgraphSyncer{
		cntr: netCntr,
		cfg:  cfg,
	}
}

const SUBGRAPH_POLL_SECS = 1

func (s *SubgraphSyncer) pollSubgraphUpdates() {
	for true {
		time.Sleep(time.Second)
		hasMore, _ := s.checkNewSubgraphSync()
		if hasMore {
			log.Printf("New subgraph time %d", s.lastSyncTime)
		}
	}
}

func (s *SubgraphSyncer) checkNewSubgraphSync() (bool, error) {
	metaTime, err := loader.LatestSubgraphTime(s.cfg)
	if err != nil {
		log.Println("Warning unable to sync subgraph meta query" + err.Error())
		return false, err
	}

	if metaTime > s.lastSyncTime {
		s.syncStep(metaTime)
		return true, nil
	}
	return false, nil
}

func (s *SubgraphSyncer) syncStart() {
	syncTime, err := loader.LatestSubgraphTime(s.cfg)

	if err != nil {
		log.Fatal("Subgraph not responding for initital sync time")
	}

	s.syncStep(syncTime)
	s.pollSubgraphUpdates()
}

func (s *SubgraphSyncer) syncStep(syncTime int) {
	startTime := s.lastSyncTime + 1

	s.cfg.Query = "./artifacts/graphQueries/balances.query"
	tblBal := tables.BalanceTable{}
	syncBal := loader.NewSyncChannel[tables.Balance, tables.BalanceSubGraph](
		tblBal, s.cfg, s.cntr.IngestBalance)
	nRows, _ := syncBal.SyncTableToSubgraph(false, startTime, syncTime)
	if nRows > 0 {
		log.Println("Sync UserBalance subgraph with rows=", nRows)
	}

	s.cfg.Query = "./artifacts/graphQueries/liqchanges.query"
	tblLiq := tables.LiqChangeTable{}
	syncLiq := loader.NewSyncChannel[tables.LiqChange, tables.LiqChangeSubGraph](
		tblLiq, s.cfg, s.cntr.IngestLiqChange)
	nRows, _ = syncLiq.SyncTableToSubgraph(false, startTime, syncTime)
	if nRows > 0 {
		log.Println("Sync LiqChanges subgraph with rows=", nRows)
	}

	s.cfg.Query = "./artifacts/graphQueries/swaps.query"
	tblSwap := tables.SwapsTable{}
	syncSwap := loader.NewSyncChannel[tables.Swap, tables.SwapSubGraph](
		tblSwap, s.cfg, s.cntr.IngestSwap)
	nRows, _ = syncSwap.SyncTableToSubgraph(false, startTime, syncTime)
	if nRows > 0 {
		log.Println("Sync Swaps subgraph with rows=", nRows)
	}

	s.cfg.Query = "./artifacts/graphQueries/knockoutcrosses.query"
	tblKo := tables.KnockoutTable{}
	syncKo := loader.NewSyncChannel[tables.KnockoutCross, tables.KnockoutCrossSubGraph](
		tblKo, s.cfg, s.cntr.IngestKnockout)
	nRows, _ = syncKo.SyncTableToSubgraph(false, startTime, syncTime)
	if nRows > 0 {
		log.Println("Sync Knockout subgraph with rows=", nRows)
	}

	s.cfg.Query = "./artifacts/graphQueries/feechanges.query"
	tblFee := tables.FeeTable{}
	syncFee := loader.NewSyncChannel[tables.FeeChange, tables.FeeChangeSubGraph](
		tblFee, s.cfg, s.cntr.IngestFee)
	nRows, _ = syncFee.SyncTableToSubgraph(false, startTime, syncTime)
	if nRows > 0 {
		log.Println("Sync FeeChanges subgraph with rows=", nRows)
	}

	s.lastSyncTime = syncTime
}

func (s *SubgraphSyncer) syncPricingSwaps() {
	s.cfg.Query = "./artifacts/graphQueries/swaps.query"
	tbl := tables.SwapsTable{}
	sync := loader.NewSyncChannel[tables.Swap, tables.SwapSubGraph](
		tbl, s.cfg, s.cntr.IngestSwap)

	LOOKBACK_WINDOW := 3600 * 1
	endTime := int(time.Now().Unix())
	startTime := endTime - LOOKBACK_WINDOW

	nRows, _ := sync.SyncTableToSubgraph(false, startTime, endTime)
	log.Println("Sync Pricing swaps subgraph with rows=", nRows)
}
