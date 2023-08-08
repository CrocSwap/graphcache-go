package main

import (
	"flag"
	"strconv"
	"time"

	"github.com/CrocSwap/graphcache-go/cache"
	"github.com/CrocSwap/graphcache-go/controller"
	"github.com/CrocSwap/graphcache-go/db"
	"github.com/CrocSwap/graphcache-go/loader"
	"github.com/CrocSwap/graphcache-go/server"
	"github.com/CrocSwap/graphcache-go/utils"
	"github.com/CrocSwap/graphcache-go/views"
)

var uniswapCandles = utils.GoDotEnvVariable("UNISWAP_CANDLES") == "true"
func main() {
	var netCfgPath = flag.String("netCfg", "./config/networks.json", "network config file")
	if(uniswapCandles){
		*netCfgPath = "./config/uniswapNetwork.json"
	}
	flag.Parse()

	netCfg := loader.LoadNetworkConfig(*netCfgPath)
	onChain := loader.OnChainLoader{Cfg: netCfg}

	cache := cache.New()
	cntrl := controller.New(netCfg, cache)
	for network, chainCfg := range netCfg {
		startTime :=  int(time.Now().Unix())
		controller.NewSubgraphSyncer(cntrl, chainCfg, network, startTime)
		if(uniswapCandles){
			hourToSyncUniswapShards, err := strconv.Atoi(utils.GoDotEnvVariable("UNISWAP_HOUR_TO_SYNC_SHARDS"))
			if err != nil {
				hourToSyncUniswapShards = 1
			}
			gmt := time.FixedZone("GMT", 0)
			time.Local = gmt
		
			go db.ScheduleSyncShards(hourToSyncUniswapShards, chainCfg)
			controller.NewUniswapSyncer(cntrl, chainCfg, network, startTime)
		}

	}

	views := views.Views{Cache: cache, OnChain: &onChain}
	apiServer := server.APIWebServer{Views: &views}
	apiServer.Serve()
}
