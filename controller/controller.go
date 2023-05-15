package controller

import (
	"log"

	"github.com/CrocSwap/graphcache-go/loader"
	"github.com/CrocSwap/graphcache-go/models"
	"github.com/CrocSwap/graphcache-go/tables"
	"github.com/CrocSwap/graphcache-go/types"
)

type Controller struct {
	netCfg loader.NetworkConfig
	models *models.Models
}

func New(netCfg loader.NetworkConfig, models *models.Models) *Controller {
	return &Controller{
		netCfg: netCfg,
		models: models,
	}
}

type ControllerOverNetwork struct {
	chainId types.ChainId
	ctrl    *Controller
}

func (c *Controller) OnNetwork(network types.NetworkName) *ControllerOverNetwork {
	chainId, okay := c.netCfg.ChainIDForNetwork(network)
	if !okay {
		log.Fatal("No network config for " + network)
	}

	return &ControllerOverNetwork{
		ctrl:    c,
		chainId: chainId,
	}
}

func (c *ControllerOverNetwork) IngestBalance(b tables.Balance) {
	token := types.RequireEthAddr(b.Token)
	user := types.RequireEthAddr(b.User)
	c.ctrl.models.AddUserBalance(c.chainId, user, token)
}
