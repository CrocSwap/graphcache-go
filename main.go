package main

import (
	"flag"

	"github.com/CrocSwap/graphcache-go/cache"
	"github.com/CrocSwap/graphcache-go/controller"
	"github.com/CrocSwap/graphcache-go/loader"
	"github.com/CrocSwap/graphcache-go/server"
	"github.com/CrocSwap/graphcache-go/views"
)

func main() {
	var netCfgPath = flag.String("netCfg", "./config/networks.json", "network config file")
	var apiPath = flag.String("apiPath", "gcgo", "API server root path")
	var noRpcMode = flag.Bool("noRpcMode", false, "Run in mode with no RPC calls")
	var swapStart = flag.Int("swapStart", 0, "Block number to start swap event processing")
	var aggStart = flag.Int("aggStart", 0, "Block number to start aggregate event processing")
	var balStart = flag.Int("balStart", 0, "Block number to start user balance processing")

	flag.Parse()

	netCfg := loader.LoadNetworkConfig(*netCfgPath)
	onChain := loader.NewOnChainLoader(netCfg)

	cache := cache.New()
	cntrl := controller.New(netCfg, cache, onChain)

	if *noRpcMode {
		nonQuery := loader.NonCrocQuery{}
		cntrl = controller.NewOnQuery(netCfg, cache, &nonQuery)
	}

	syncs := make([]*controller.SubgraphSyncer, 0)

	for network, chainCfg := range netCfg {
		startBlocks := controller.SubgraphStartBlocks{
			Swaps: *swapStart,
			Aggs:  *aggStart,
			Bal:   *balStart,
		}
		syncer := controller.NewSubgraphSyncerAtStart(cntrl, chainCfg, network, startBlocks)
		syncs = append(syncs, syncer)
	}

	cntrl.SpinUntilLiqSync()
	for _, syncer := range syncs {
		go syncer.PollSubgraphUpdates()
	}

	views := views.Views{Cache: cache, OnChain: onChain}
	apiServer := server.APIWebServer{Views: &views}
	apiServer.Serve(*apiPath)
}
