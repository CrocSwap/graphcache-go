package main

import (
	"github.com/CrocSwap/graphcache-go/controller"
	"github.com/CrocSwap/graphcache-go/loader"
	"github.com/CrocSwap/graphcache-go/models"
	"github.com/CrocSwap/graphcache-go/server"
	"github.com/CrocSwap/graphcache-go/types"
	"github.com/CrocSwap/graphcache-go/views"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	netCfgPath := "../graphcache/webserver/config/networks.json"
	netCfg := loader.LoadNetworkConfig(netCfgPath)
	models := models.New()
	controller := controller.New(netCfg, models)

	goerlChainConfig, _ := netCfg["goerli"]
	goerliCntrl := controller.OnNetwork(types.NetworkName("goerli"))
	cfg := loader.SyncChannelConfig{
		Chain:   goerlChainConfig,
		Network: "goerli",
		Query:   "../graphcache/webserver/queries/balances.query",
	}
	sync := loader.NewSyncChannel(cfg, goerliCntrl.IngestBalance)

	sync.SyncTableFromDb("../_data/database.db")
	sync.SyncTableToSubgraph()

	views := views.Views{Models: models}
	apiServer := server.APIWebServer{Views: &views}
	apiServer.Serve()
}
