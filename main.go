package main

import (
	"flag"
	"fmt"
	"runtime/metrics"

	"github.com/CrocSwap/graphcache-go/cache"
	"github.com/CrocSwap/graphcache-go/controller"
	"github.com/CrocSwap/graphcache-go/loader"
	"github.com/CrocSwap/graphcache-go/server"
	"github.com/CrocSwap/graphcache-go/views"
)

func getMemoryLimit() {
	const myMetric = "/gc/gomemlimit:bytes"
	sample := make([]metrics.Sample, 1)
	sample[0].Name = myMetric
	metrics.Read(sample)
	if sample[0].Value.Kind() == metrics.KindBad {
		panic(fmt.Sprintf("metric %q no longer supported", myMetric))
	}
	freeBytes := sample[0].Value.Uint64()
	fmt.Printf("memlimit: %d\n", freeBytes)
}

func main() {
	// log.SetFlags(log.LstdFlags | log.Lshortfile)
	getMemoryLimit()
	var netCfgPath = flag.String("netCfg", "./config/ethereum.json", "network config file")
	var apiPath = flag.String("apiPath", "gcgo", "API server root path")
	var listenAddr = flag.String("listenAddr", ":8080", "HTTP server listen address")
	var noRpcMode = flag.Bool("noRpcMode", false, "Run in mode with no RPC calls")
	var swapStart = flag.Int("swapStart", 0, "Block number to start swap event processing")
	var aggStart = flag.Int("aggStart", 0, "Block number to start aggregate event processing")
	var balStart = flag.Int("balStart", 0, "Block number to start user balance processing")
	var extendedApi = flag.Bool("extendedApi", false, "Expose additional methods in the API")
	var combinedQuery = flag.Bool("combinedQuery", false, "Use the combined subgraph query instead of individual ones")
	var startupCache = flag.String("startupCache", "", "Either directory or HTTP URL to load startup cache from")

	flag.Parse()

	netCfg := loader.LoadNetworkConfig(*netCfgPath)
	onChain := loader.NewOnChainLoader(netCfg)

	cache := cache.New()
	cntrl := controller.New(netCfg, cache, onChain)

	if *noRpcMode {
		nonQuery := loader.NonCrocQuery{}
		cntrl = controller.NewOnQuery(netCfg, cache, &nonQuery)
	}

	syncs := make([]controller.SubgraphSyncer, 0)

	for network, chainCfg := range netCfg {
		startBlocks := loader.SubgraphStartBlocks{
			Swaps: *swapStart,
			Aggs:  *aggStart,
			Bal:   *balStart,
		}
		var syncer controller.SubgraphSyncer
		if *combinedQuery {
			syncer = controller.NewCombinedSubgraphSyncerAtStart(cntrl, chainCfg, network, startBlocks, *startupCache)
		} else {
			syncer = controller.NewSubgraphSyncerAtStart(cntrl, chainCfg, network, startBlocks, *startupCache)
		}
		syncs = append(syncs, syncer)
	}

	if *noRpcMode == false {
		cntrl.StartupSubgraphSyncDone()
	}

	for _, syncer := range syncs {
		go syncer.PollSubgraphUpdates()
	}

	views := views.Views{Cache: cache, OnChain: onChain}
	apiServer := server.APIWebServer{Views: &views}
	apiServer.Serve(*apiPath, *listenAddr, *extendedApi)
}
