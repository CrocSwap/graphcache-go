package controller

import (
	"log"
	"time"

	"github.com/CrocSwap/graphcache-go/loader"
	"github.com/CrocSwap/graphcache-go/tables"
	"github.com/CrocSwap/graphcache-go/types"
)

type SubgraphSyncer struct {
	cntr *ControllerOverNetwork
	cfg  loader.SyncChannelConfig
}

func NewSubgraphSyncer(controller *Controller, chainConfig loader.ChainConfig, network types.NetworkName) SubgraphSyncer {
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

func (s *SubgraphSyncer) SyncStartup(chainConfig loader.ChainConfig, network types.NetworkName) {
	s.cfg.Query = "./artifacts/graphQueries/balances.query"
	tblBal := tables.BalanceTable{}
	syncBal := loader.NewSyncChannel[tables.Balance, tables.BalanceSubGraph](
		tblBal, s.cfg, s.cntr.IngestBalance)
	nRows, _ := syncBal.SyncTableToSubgraph(false)
	log.Println("Sync UserBalance subgraph with rows=", nRows)

	s.cfg.Query = "./artifacts/graphQueries/liqchanges.query"
	tblLiq := tables.LiqChangeTable{}
	syncLiq := loader.NewSyncChannel[tables.LiqChange, tables.LiqChangeSubGraph](
		tblLiq, s.cfg, s.cntr.IngestLiqChange)
	nRows, _ = syncLiq.SyncTableToSubgraph(false)
	log.Println("Sync LiqChanges subgraph with rows=", nRows)

	s.cfg.Query = "./artifacts/graphQueries/swaps.query"
	tblSwap := tables.SwapsTable{}
	syncSwap := loader.NewSyncChannel[tables.Swap, tables.SwapSubGraph](
		tblSwap, s.cfg, s.cntr.IngestSwap)
	nRows, _ = syncSwap.SyncTableToSubgraph(false)
	log.Println("Sync Swaps subgraph with rows=", nRows)

	s.cfg.Query = "./artifacts/graphQueries/knockoutcrosses.query"
	tblKo := tables.KnockoutTable{}
	syncKo := loader.NewSyncChannel[tables.KnockoutCross, tables.KnockoutCrossSubGraph](
		tblKo, s.cfg, s.cntr.IngestKnockout)
	nRows, _ = syncKo.SyncTableToSubgraph(false)
	log.Println("Sync Knockout subgraph with rows=", nRows)

	s.cfg.Query = "./artifacts/graphQueries/feechanges.query"
	tblFee := tables.FeeTable{}
	syncFee := loader.NewSyncChannel[tables.FeeChange, tables.FeeChangeSubGraph](
		tblFee, s.cfg, s.cntr.IngestFee)
	nRows, _ = syncFee.SyncTableToSubgraph(false)
	log.Println("Sync FeeChanges subgraph with rows=", nRows)
}

func (s *SubgraphSyncer) SyncPricingSwaps() {
	s.cfg.Query = "./artifacts/graphQueries/swaps.query"
	tbl := tables.SwapsTable{}
	sync := loader.NewSyncChannel[tables.Swap, tables.SwapSubGraph](
		tbl, s.cfg, s.cntr.IngestSwap)

	LOOKBACK_WINDOW := 3600 * 1
	sync.LastObserved = int(time.Now().Unix()) - LOOKBACK_WINDOW

	nRows, _ := sync.SyncTableToSubgraph(true)
	log.Println("Sync Pricing swaps subgraph with rows=", nRows)
}
