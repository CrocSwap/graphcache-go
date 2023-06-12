package main

import (
	"fmt"

	"github.com/CrocSwap/graphcache-go/cache"
	"github.com/CrocSwap/graphcache-go/controller"
	"github.com/CrocSwap/graphcache-go/loader"
	"github.com/CrocSwap/graphcache-go/server"
	"github.com/CrocSwap/graphcache-go/views"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	netCfgPath := "./config/networks.json"
	netCfg := loader.LoadNetworkConfig(netCfgPath)
	onChain := loader.OnChainLoader{Cfg: netCfg}

	cache := cache.New()
	cntrl := controller.New(netCfg, cache)
	fmt.Println("A")

	goerlChainConfig, _ := netCfg["goerli"]
	controller.NewSubgraphSyncer(cntrl, goerlChainConfig, "goerli")

	mainnetChainConfig, _ := netCfg["mainnet"]
	controller.NewSubgraphSyncer(cntrl, mainnetChainConfig, "mainnet")

	fmt.Println("B")

	views := views.Views{Cache: cache, OnChain: &onChain}
	apiServer := server.APIWebServer{Views: &views}
	apiServer.Serve()
}
