package controller

import (
	"log"
	"time"

	"github.com/CrocSwap/graphcache-go/loader"
	"github.com/CrocSwap/graphcache-go/tables"
	"github.com/CrocSwap/graphcache-go/types"
)

func (c *Controller) SyncSubgraph(chainConfig loader.ChainConfig, network types.NetworkName) {
	cfg := loader.SyncChannelConfig{
		Chain:   chainConfig,
		Network: network,
	}
	netCntr := c.OnNetwork(network)

	cfg.Query = "./artifacts/graphQueries/balances.query"
	tblBal := tables.BalanceTable{}
	syncBal := loader.NewSyncChannel[tables.Balance, tables.BalanceSubGraph](
		tblBal, cfg, netCntr.IngestBalance)
	nRows, _ := syncBal.SyncTableToSubgraph(false)
	log.Println("Sync UserBalance subgraph with rows=", nRows)

	cfg.Query = "./artifacts/graphQueries/liqchanges.query"
	tblLiq := tables.LiqChangeTable{}
	syncLiq := loader.NewSyncChannel[tables.LiqChange, tables.LiqChangeSubGraph](
		tblLiq, cfg, netCntr.IngestLiqChange)
	nRows, _ = syncLiq.SyncTableToSubgraph(false)
	log.Println("Sync LiqChanges subgraph with rows=", nRows)

	cfg.Query = "./artifacts/graphQueries/swaps.query"
	tblSwap := tables.SwapsTable{}
	syncSwap := loader.NewSyncChannel[tables.Swap, tables.SwapSubGraph](
		tblSwap, cfg, netCntr.IngestSwap)
	nRows, _ = syncSwap.SyncTableToSubgraph(false)
	log.Println("Sync Swaps subgraph with rows=", nRows)

	cfg.Query = "./artifacts/graphQueries/knockoutcrosses.query"
	tblKo := tables.KnockoutTable{}
	syncKo := loader.NewSyncChannel[tables.KnockoutCross, tables.KnockoutCrossSubGraph](
		tblKo, cfg, netCntr.IngestKnockout)
	nRows, _ = syncKo.SyncTableToSubgraph(false)
	log.Println("Sync Knockout subgraph with rows=", nRows)

	cfg.Query = "./artifacts/graphQueries/feechanges.query"
	tblFee := tables.FeeTable{}
	syncFee := loader.NewSyncChannel[tables.FeeChange, tables.FeeChangeSubGraph](
		tblFee, cfg, netCntr.IngestFee)
	nRows, _ = syncFee.SyncTableToSubgraph(false)
	log.Println("Sync FeeChanges subgraph with rows=", nRows)
}

func (c *Controller) SyncPricingSwaps(chainConfig loader.ChainConfig, network types.NetworkName) {
	cfg := loader.SyncChannelConfig{
		Chain:   chainConfig,
		Network: network,
	}
	netCntr := c.OnNetwork(network)

	cfg.Query = "./artifacts/graphQueries/swaps.query"
	tbl := tables.SwapsTable{}
	sync := loader.NewSyncChannel[tables.Swap, tables.SwapSubGraph](
		tbl, cfg, netCntr.IngestSwap)

	LOOKBACK_WINDOW := 3600 * 1
	sync.LastObserved = int(time.Now().Unix()) - LOOKBACK_WINDOW

	nRows, _ := sync.SyncTableToSubgraph(true)
	log.Println("Sync Pricing swaps subgraph with rows=", nRows)
}
