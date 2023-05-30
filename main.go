package main

import (
	"github.com/CrocSwap/graphcache-go/cache"
	"github.com/CrocSwap/graphcache-go/controller"
	"github.com/CrocSwap/graphcache-go/loader"
	"github.com/CrocSwap/graphcache-go/server"
	"github.com/CrocSwap/graphcache-go/views"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	netCfgPath := "../graphcache/webserver/config/networks.json"
	netCfg := loader.LoadNetworkConfig(netCfgPath)
	cache := cache.New()
	onChain := loader.OnChainLoader{Cfg: netCfg}

	goerlChainConfig, _ := netCfg["goerli"]
	controller := controller.New(netCfg, cache)
	controller.SyncSubgraph(goerlChainConfig, "goerli")

	mainnetChainConfig, _ := netCfg["mainnet"]
	controller.SyncPricingSwaps(mainnetChainConfig, "mainnet")

	views := views.Views{Cache: cache, OnChain: &onChain}
	apiServer := server.APIWebServer{Views: &views}
	apiServer.Serve()
}
