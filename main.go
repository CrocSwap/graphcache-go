package main

import (
	"fmt"

	"github.com/CrocSwap/graphcache-go/models"
	"github.com/CrocSwap/graphcache-go/server"
	"github.com/CrocSwap/graphcache-go/tables"
	"github.com/CrocSwap/graphcache-go/views"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db := openSqliteDb("../_data/database.db")
	tables.LoadTokenBalancesSql(db, func(b tables.Balance) { fmt.Println(b.Token) })

	models := models.New()
	views := views.Views{Models: models}
	apiServer := server.APIWebServer{Views: &views}
	apiServer.Serve()
}
