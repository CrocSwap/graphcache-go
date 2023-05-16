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

func (c *ControllerOverNetwork) IngestLiqChange(l tables.LiqChange) {
	liq := formLiqLoc(l)
	pool := types.PoolLocation{
		ChainId: c.chainId,
		PoolIdx: l.PoolIdx,
		Base:    types.RequireEthAddr(l.Base),
		Quote:   types.RequireEthAddr(l.Quote),
	}
	pos := types.PositionLocation{
		PoolLocation:      pool,
		LiquidityLocation: liq,
		User:              types.RequireEthAddr(l.User),
	}

	if l.ChangeType == "mint" {
		c.ctrl.models.UpdatePositionMint(pos, l.Time)
	} else if l.ChangeType == "burn" {
		c.ctrl.models.UpdatePositionBurn(pos, l.Time)
	} else if l.ChangeType == "harvest" {
		c.ctrl.models.UpdatePositionHarvest(pos, l.Time)
	}
}

func formLiqLoc(l tables.LiqChange) types.LiquidityLocation {
	if l.PositionType == "ambient" {
		return types.AmbientLiquidityLocation()
	} else if l.PositionType == "concentrated" {
		return types.ConcLiquidityLocation(l.BidTick, l.AskTick)
	} else {
		pivotTime := 0
		if l.PivotTime != nil {
			pivotTime = *l.PivotTime
		}
		return types.KnockoutLiquidityLocation(l.BidTick, l.AskTick, pivotTime, l.IsBid > 0)
	}
}
