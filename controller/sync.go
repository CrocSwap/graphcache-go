package controller

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/CrocSwap/graphcache-go/loader"
	"github.com/CrocSwap/graphcache-go/tables"
	"github.com/CrocSwap/graphcache-go/types"
	"github.com/CrocSwap/graphcache-go/utils"
)

var uniswapCandles = utils.GoDotEnvVariable("UNISWAP_CANDLES") == "true"
var testnet = utils.GoDotEnvVariable("TESTNET") == "true"

type SubgraphSyncer struct {
	cntr         *ControllerOverNetwork
	cfg          loader.SyncChannelConfig
	lastSyncTime int
}

func NewSubgraphSyncer(controller *Controller, chainConfig loader.ChainConfig, network types.NetworkName, startTime int) SubgraphSyncer {
	sync := makeSubgraphSyncer(controller, chainConfig, network)
	syncNotif := make(chan bool, 1)

	log.Printf("[Polling Syncer]: Starting poll sync for %s\n", network)
	go sync.syncStart(syncNotif, startTime)

	<-syncNotif
	return sync
}

func NewUniswapSyncer(controller *Controller, chainConfig loader.ChainConfig, network types.NetworkName, startTime int) SubgraphSyncer {
	sync := makeSubgraphSyncer(controller, chainConfig, network)
	syncNotif := make(chan bool, 1)
	go sync.historicalSyncCandles(syncNotif, startTime)
	<-syncNotif
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
			log.Printf("[Polling Syncer]: New subgraph time %s", time.Unix(int64(s.lastSyncTime), 0))
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
		s.syncStep(metaTime)
		return true, nil
	}
	return false, nil
}

func (s *SubgraphSyncer) historicalSyncCandles(notif chan bool, startTime int) {
	notif <- true // Signal that the server is ready to accept requests
	// UNISWAP_CANDLE_LOOKBACK_WINDOW :=  int(time.Now().Unix()) -  3600 *24 *7
	initialSyncTimeEnv := utils.GoDotEnvVariable("UNISWAP_CANDLE_BEGINNING_TIMESTAMP")
	testNetCandles := utils.GoDotEnvVariable("TESTNET") == "true"

	initialSyncTime, err := strconv.Atoi(initialSyncTimeEnv)
	if err != nil {
		log.Println("Falling back to default initial sync time of Jan 1 2023")
		initialSyncTime = 1672531200 /// Fall back to Jan 1 2023
	}
	if testNetCandles {
		log.Printf("[Historical Syncer]: Syncing testnet back to %s\n", time.Unix(int64(initialSyncTime), 0))

		s.syncCrocSwapCandles(false, initialSyncTime, startTime)
	} else {
		// Goes in forward from last swap time until today
		latestSwapTime, err := loader.GetLatestSwapTime()
		nextSwapTime := int(latestSwapTime) + 1
		if err != nil {
			log.Fatalf("Database not responding")
			return
		}
		log.Printf("[Historical Syncer]: Latest swap time is %s\n", time.Unix(int64(nextSwapTime), 0))

		s.syncUniswapCandles("save", nextSwapTime, startTime)
		log.Printf("[Historical Syncer]: Synced and saved to db uniswap swaps from subgraph %s to %s\n", time.Unix(int64(latestSwapTime), 0), time.Unix(int64(startTime), 0))
		s.syncUniswapCandles("db", initialSyncTime, nextSwapTime)
		log.Printf("[Historical Syncer]: Synced uniswap swaps from db from %s to %s\n", time.Unix(int64(initialSyncTime), 0), time.Unix(int64(nextSwapTime), 0))

	}

}

func (s *SubgraphSyncer) syncStart(notif chan bool, startTime int) {
	if !uniswapCandles {
		syncTime, err := loader.LatestSubgraphTime(s.cfg)
		if err != nil {
			log.Fatalf("[Polling Syncer]: Subgraph not responding from %s", s.cntr.chainCfg.Subgraph)
		}

		s.syncStep(syncTime)
	} else {
		s.cntr.FlushSyncCycle(startTime)
		s.lastSyncTime = startTime
	}

	log.Printf("[Polling Syncer]: Startup subgraph sync done on chainId=%d", s.cntr.chainCfg.ChainID)
	notif <- true

	s.pollSubgraphUpdates()
}

func (s *SubgraphSyncer) logSyncCycle(table string, nRows int) {
	if nRows > 0 {
		log.Printf("[Polling Syncer]: Sync %s subgraph on chainId=%d with rows=%d", table, s.cntr.chainCfg.ChainID, nRows)
	}
}

func (s *SubgraphSyncer) syncStep(syncTime int) {
	startTime := s.lastSyncTime + 1
	doSyncFwd := true

	if uniswapCandles {
		if testnet {
			s.syncCrocSwapCandles(doSyncFwd, startTime, syncTime)
		} else {
			s.syncUniswapCandles("subgraph", startTime, syncTime)
		}
	} else {

		s.cfg.Query = "./artifacts/graphQueries/balances.query"
		tblBal := tables.BalanceTable{}
		syncBal := loader.NewSyncChannel[tables.Balance, tables.BalanceSubGraph](
			tblBal, s.cfg, s.cntr.IngestBalance, nil)
		nRows, _ := syncBal.SyncTableToSubgraph(doSyncFwd, startTime, syncTime)
		s.logSyncCycle("User Balances", nRows)

		s.cfg.Query = "./artifacts/graphQueries/liqchanges.query"
		tblLiq := tables.LiqChangeTable{}
		syncLiq := loader.NewSyncChannel[tables.LiqChange, tables.LiqChangeSubGraph](
			tblLiq, s.cfg, s.cntr.IngestLiqChange, nil)
		nRows, _ = syncLiq.SyncTableToSubgraph(doSyncFwd, startTime, syncTime)
		s.logSyncCycle("LiqChanges", nRows)

		s.cfg.Query = "./artifacts/graphQueries/swaps.query"
		tblSwap := tables.SwapsTable{}
		syncSwap := loader.NewSyncChannel[tables.Swap, tables.SwapSubGraph](
			tblSwap, s.cfg, s.cntr.IngestSwap, nil)
		nRows, _ = syncSwap.SyncTableToSubgraph(doSyncFwd, startTime, syncTime)
		s.logSyncCycle("Swaps", nRows)

		s.cfg.Query = "./artifacts/graphQueries/knockoutcrosses.query"
		tblKo := tables.KnockoutTable{}
		syncKo := loader.NewSyncChannel[tables.KnockoutCross, tables.KnockoutCrossSubGraph](
			tblKo, s.cfg, s.cntr.IngestKnockout, nil)
		nRows, _ = syncKo.SyncTableToSubgraph(doSyncFwd, startTime, syncTime)
		s.logSyncCycle("Knockout crosses", nRows)

		s.cfg.Query = "./artifacts/graphQueries/feechanges.query"
		tblFee := tables.FeeTable{}
		syncFee := loader.NewSyncChannel[tables.FeeChange, tables.FeeChangeSubGraph](
			tblFee, s.cfg, s.cntr.IngestFee, nil)
		nRows, _ = syncFee.SyncTableToSubgraph(doSyncFwd, startTime, syncTime)
		s.logSyncCycle("Fee Changes", nRows)

		s.cfg.Query = "./artifacts/graphQueries/aggevent.query"
		tblAgg := tables.AggEventsTable{}
		syncAgg := loader.NewSyncChannel[tables.AggEvent, tables.AggEventSubGraph](
			tblAgg, s.cfg, s.cntr.IngestAggEvent, nil)
		nRows, _ = syncAgg.SyncTableToSubgraph(doSyncFwd, startTime, syncTime)
		s.logSyncCycle("Poll Agg Events", nRows)
	}
	s.cntr.FlushSyncCycle(syncTime)
	s.lastSyncTime = syncTime
}

func (s *SubgraphSyncer) syncCrocSwapCandles(doSyncFwd bool, startTime int, syncTime int) {
	s.cfg.Query = "./artifacts/graphQueries/swaps.query"
	tbl := tables.Croc91SwapsTable{}
	sync := loader.NewSyncChannel[tables.AggEvent, tables.SwapSubGraph](
		tbl, s.cfg, s.cntr.IngestAggEvent, nil)
	nRows, _ := sync.SyncTableToSubgraph(doSyncFwd, startTime, syncTime)
	s.logSyncCycle("Poll Agg Events", nRows)
}
func (s *SubgraphSyncer) syncUniswapCandles(action string, startTime int, syncTime int) {
	s.cfg.Query = "./artifacts/graphQueries/swaps.uniswap.query"
	s.cfg.Chain.Subgraph = "https://api.thegraph.com/subgraphs/name/uniswap/uniswap-v3"
	tblAgg := tables.UniSwapsTable{}
	var nRows int

	switch action {
	case "subgraph":
		syncAgg := loader.NewSyncChannel[tables.AggEvent, tables.UniSwapSubGraph](
			tblAgg, s.cfg, s.cntr.IngestAggEvent, nil)
		nRows, _ = syncAgg.SyncTableToSubgraph(true, startTime, syncTime)
	case "db":
		syncAgg := loader.NewSyncChannel[tables.AggEvent, tables.UniSwapSubGraph](
			tblAgg, s.cfg, s.cntr.IngestAggEvent, nil)
		nRows, _ = syncAgg.SyncTableToDB(false, startTime, syncTime)
	case "save":
		syncAgg := loader.NewSyncChannel[tables.AggEvent, tables.UniSwapSubGraph](
			tblAgg, s.cfg, s.cntr.IngestAggEvent, loader.SaveAggEventSubgraphArrayToDB)
		nRows, _ = syncAgg.SyncTableToSubgraph(true, startTime, syncTime)

	default:
		fmt.Println("Invalid action:", action)
		return
	}
	s.logSyncCycle("Poll Agg Events", nRows)
}

func (s *SubgraphSyncer) syncPricingSwaps() {
	s.cfg.Query = "./artifacts/graphQueries/swaps.query"
	tbl := tables.SwapsTable{}
	sync := loader.NewSyncChannel[tables.Swap, tables.SwapSubGraph](
		tbl, s.cfg, s.cntr.IngestSwap, nil)

	LOOKBACK_WINDOW := 3600 * 1
	endTime := int(time.Now().Unix())
	startTime := endTime - LOOKBACK_WINDOW

	nRows, _ := sync.SyncTableToSubgraph(false, startTime, endTime)
	log.Println("Sync Pricing swaps subgraph with rows=", nRows)
}
