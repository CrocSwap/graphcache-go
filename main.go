package main

import (
	"flag"

	"github.com/CrocSwap/graphcache-go/cache"
	"github.com/CrocSwap/graphcache-go/controller"
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
		controller.NewSubgraphSyncer(cntrl, chainCfg, network, false)
		if(uniswapCandles){
			controller.NewSubgraphSyncer(cntrl, chainCfg, network, true)
		}
	}

	views := views.Views{Cache: cache, OnChain: &onChain}
	apiServer := server.APIWebServer{Views: &views}
	apiServer.Serve()
}
