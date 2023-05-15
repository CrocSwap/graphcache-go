package main

import (
	"github.com/CrocSwap/graphcache-go/models"
	"github.com/CrocSwap/graphcache-go/server"
	"github.com/CrocSwap/graphcache-go/views"
)

func main() {
	models := models.New()
	views := views.Views{Models: models}
	apiServer := server.APIWebServer{Views: &views}
	apiServer.Serve()
}
