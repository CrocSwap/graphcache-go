package controller

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/CrocSwap/graphcache-go/db"
	"github.com/CrocSwap/graphcache-go/loader"
	"github.com/CrocSwap/graphcache-go/tables"
	"github.com/CrocSwap/graphcache-go/types"
	"github.com/CrocSwap/graphcache-go/utils"
)

var uniswapCandles = utils.GoDotEnvVariable("UNISWAP_CANDLES") == "true" 

type IngestionItem  struct {
	Name string
	Path string
	Method string // local, gcs, subgraph 
	// TODO: better typing on method
};

func NewUniswapSyncer(controller *Controller, chainConfig loader.ChainConfig, network types.NetworkName, serverStartupTime int) SubgraphSyncer {
	sync := makeSubgraphSyncer(controller, chainConfig, network)
	syncNotif := make(chan bool, 1)
	go sync.historicalSyncCandles(syncNotif, serverStartupTime)
	<-syncNotif
	return sync
}



func createIngestionList() []IngestionItem{
	GCSShards, err := db.FetchBucketItems(db.BucketName)

	if err != nil {
		log.Println("[Shard Syncer]: Error ", err)
		return nil
	}

	neededShards := db.GetDaysList(db.InitialTimestamp)
	var ingestionList []IngestionItem
	for _, shardName := range neededShards {
		shardPath := fmt.Sprintf("%s/%s", db.ShardsPath, shardName)
		fullShardPath := shardPath + ".db"
		shardData := IngestionItem{
			Name:   shardName,
			Path:   fullShardPath,
		}
		if(db.FileExistsInDir(fullShardPath)){
			shardData.Method = "local"
		} else if db.FilePathExistsInBucket(fullShardPath, GCSShards){
			shardData.Method = "gcs"
		} else {
			shardData.Method = "subgraph"
		}
		ingestionList = append(ingestionList, shardData)

	}

	return ingestionList

}
func (s *SubgraphSyncer) historicalSyncCandles(notif chan bool, serverStartupTime int) {
	// notif <- true // Signal that the server is ready to accept requests
	// UNISWAP_CANDLE_LOOKBACK_WINDOW :=  int(time.Now().Unix()) -  3600 *24 *7
	// TODO: remove this env var
	daysOfCandlesBeforeServerReady, err := strconv.Atoi(utils.GoDotEnvVariable("DAYS_OF_UNISWAP_CANDLES_BEFORE_SERVER_READY"))
	if err != nil {
		daysOfCandlesBeforeServerReady = 0
	}
	
	currentTime :=  time.Now()
    startOfToday := currentTime.Truncate(24 * time.Hour).Unix()

	ingestionList := createIngestionList()
	log.Printf("[Historical Syncer]: now syncing Uniswaps from %s to %s\n", time.Unix(int64(startOfToday), 0), time.Unix(int64(serverStartupTime), 0))

	s.syncUniswapCandles("subgraph", int(startOfToday), serverStartupTime, "")
	log.Printf("[Historical Syncer]: Synced uniswap swaps from subgraph %s to %s\n", time.Unix(int64(startOfToday), 0), time.Unix(int64(serverStartupTime), 0))


	firedNotif := false
	for i, ingestionItem := range ingestionList {
		log.Println("Ingesting item",ingestionItem.Name)
		if  i > daysOfCandlesBeforeServerReady && !firedNotif {
			log.Printf("[Historical Syncer]: Exposing API after ingesting %d days of data\n", daysOfCandlesBeforeServerReady)
			notif <- true
			firedNotif = true
		}
		switch(ingestionItem.Method){
			case "gcs":
			// First download, then sync normally
			log.Printf("[Historical Syncer]: Downloading shard from GCS Bucket: %s \n", ingestionItem.Name)
			db.DownloadShardFromBucket(ingestionItem.Path)
			case "subgraph":
			// Fetch from subgraph while saving to a shard

			dayStartTime := db.GetStartOfDayTimestamp(ingestionItem.Name)
			dayEndTime := db.GetEndOfDayTimestamp(ingestionItem.Name)
			log.Printf("[Historical Syncer]: Creating shard from Subgraph: %s between dates %s and %s \n", ingestionItem.Name,  time.Unix(int64(dayStartTime), 0),  time.Unix(int64(dayEndTime), 0))

			db.FetchUniswapAndSaveToShard(s.cfg.Chain, ingestionItem.Path, int(dayStartTime), int(dayEndTime))
			log.Printf("[Historical Syncer]: Created shard: %s \n", ingestionItem.Name)

		}
		s.syncUniswapCandles("shard", int(db.InitialTimestamp), serverStartupTime, ingestionItem.Path)
		log.Printf("[Historical Syncer]: Synced local shard %s \n", ingestionItem.Name)

	}

}


func (s *SubgraphSyncer) syncUniswapCandles(action string, startTime int, syncTime int, dbString string) {
	s.cfg.Query = "./artifacts/graphQueries/swaps.uniswap.query"
	s.cfg.Chain.Subgraph = "https://api.thegraph.com/subgraphs/name/uniswap/uniswap-v3"
	tblAgg := tables.UniSwapsTable{}
	var nRows int

	switch action {
	case "subgraph":
		syncAgg := loader.NewSyncChannel[tables.AggEvent, tables.UniSwapSubGraph](
			tblAgg, s.cfg, s.cntr.IngestAggEvent)
		nRows, _ = syncAgg.SyncTableToSubgraph(true, startTime, syncTime)
	case "shard":
		syncAgg := loader.NewSyncChannel[tables.AggEvent, tables.UniSwapSubGraph](
			tblAgg, s.cfg, s.cntr.IngestAggEvent)
		nRows, _ = syncAgg.SyncTableToDB(false, startTime, syncTime, dbString)

	default:
		fmt.Println("Invalid action:", action)
		return
	}
	s.logSyncCycle("Poll Agg Events", nRows)
}

