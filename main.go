package main

import (
	"github.com/CrocSwap/graphcache-go/cache"
	"github.com/CrocSwap/graphcache-go/controller"
	"github.com/CrocSwap/graphcache-go/loader"
	"github.com/CrocSwap/graphcache-go/server"
	"github.com/CrocSwap/graphcache-go/tables"
	"github.com/CrocSwap/graphcache-go/types"
	"github.com/CrocSwap/graphcache-go/views"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	netCfgPath := "../graphcache/webserver/config/networks.json"
	netCfg := loader.LoadNetworkConfig(netCfgPath)
	cache := cache.New()
	controller := controller.New(netCfg, cache)

	goerlChainConfig, _ := netCfg["goerli"]
	goerliCntrl := controller.OnNetwork(types.NetworkName("goerli"))
	cfg := loader.SyncChannelConfig{
		Chain:   goerlChainConfig,
		Network: "goerli",
		Query:   "../graphcache/webserver/queries/balances.query",
	}

	tbl := tables.BalanceTable{}
	sync := loader.NewSyncChannel[tables.Balance, tables.BalanceSubGraph](
		tbl, cfg, goerliCntrl.IngestBalance)

	sync.SyncTableFromDb("../_data/database.db")
	sync.SyncTableToSubgraph()

	cfg.Query = "../graphcache/webserver/queries/liqchanges.query"
	tbl2 := tables.LiqChangeTable{}
	sync2 := loader.NewSyncChannel[tables.LiqChange, tables.LiqChangeSubGraph](
		tbl2, cfg, goerliCntrl.IngestLiqChange)

	sync2.SyncTableFromDb("../_data/database.db")
	sync2.SyncTableToSubgraph()

	views := views.Views{Cache: cache}
	apiServer := server.APIWebServer{Views: &views}
	apiServer.Serve()
}
