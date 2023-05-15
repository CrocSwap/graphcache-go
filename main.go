package main

import (
	"fmt"

	"github.com/CrocSwap/graphcache-go/loader"
	"github.com/CrocSwap/graphcache-go/models"
	"github.com/CrocSwap/graphcache-go/server"
	"github.com/CrocSwap/graphcache-go/tables"
	"github.com/CrocSwap/graphcache-go/views"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	netConfig := loader.LoadChainConfigs("../graphcache/webserver/config/networks.json")

	chainConfig, _ := netConfig["goerli"]
	cfg := loader.SyncChannelConfig{
		Chain:   chainConfig,
		Network: "goerli",
		Query:   "../graphcache/webserver/queries/balances.query",
	}
	sync := loader.NewSyncChannel(cfg, func(b tables.Balance) { fmt.Println(b) })

	db := loader.OpenSqliteDb("../_data/database.db")
	sync.SyncTableFromDb(db)
	sync.SyncTableToSubgraph()

	models := models.New()
	views := views.Views{Models: models}
	apiServer := server.APIWebServer{Views: &views}
	apiServer.Serve()
}
