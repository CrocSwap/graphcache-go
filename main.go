package main

import (
	"flag"
	"fmt"

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
	flag.Parse()

	netCfg := loader.LoadNetworkConfig(*netCfgPath)
	onChain := loader.OnChainLoader{Cfg: netCfg}

	cache := cache.New()
	cntrl := controller.New(netCfg, cache)

	for network, chainCfg := range netCfg {
		// startTime :=  int(time.Now().Unix())
		// yesterday := startTime - 3600*24
		// controller.NewSubgraphSyncer(cntrl, chainCfg, network, startTime)
		// if(uniswapCandles){
		// 	// controller.NewUniswapSyncer(cntrl, chainCfg, network, startTime, "./_data/database.db")
		// 	// controller.NewUniswapSyncer(cntrl, chainCfg, network, startTime, "./_data/eth.db")
		// 	// loader.FetchUniswapAndSaveToShard(chainCfg, "./_data/new.db", yesterday, startTime)

		// }
		fmt.Println("cntrl, network", cntrl, network, chainCfg)
		// db.SyncLocalShardsWithUniswap(chainCfg)
		db.SyncLocalShardsWithUniswap(chainCfg)

	}

	views := views.Views{Cache: cache, OnChain: &onChain}
	apiServer := server.APIWebServer{Views: &views}
	apiServer.Serve()
}
