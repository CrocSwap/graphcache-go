package controller

import (
	"log"

	"github.com/CrocSwap/graphcache-go/cache"
	"github.com/CrocSwap/graphcache-go/loader"
	"github.com/CrocSwap/graphcache-go/tables"
	"github.com/CrocSwap/graphcache-go/types"
)

type Controller struct {
	netCfg  loader.NetworkConfig
	cache   *cache.MemoryCache
	workers *workers
}

func New(netCfg loader.NetworkConfig, cache *cache.MemoryCache) *Controller {
	return &Controller{
		netCfg:  netCfg,
		cache:   cache,
		workers: initWorkers(netCfg),
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
	c.ctrl.cache.AddUserBalance(c.chainId, user, token)
}

func (c *ControllerOverNetwork) IngestLiqChange(l tables.LiqChange) {
	if l.PositionType == "knockout" {
		return
	} else {
		c.ingestPassiveLiq(l)
	}
}

func (c *ControllerOverNetwork) ingestPassiveLiq(l tables.LiqChange) {
	liq := formLiqLoc(l)
	pool := types.PoolLocation{
		ChainId: c.chainId,
		PoolIdx: l.PoolIdx,
		Base:    types.RequireEthAddr(l.Base),
		Quote:   types.RequireEthAddr(l.Quote),
	}
	loc := types.PositionLocation{
		PoolLocation:      pool,
		LiquidityLocation: liq,
		User:              types.RequireEthAddr(l.User),
	}

	pos := c.ctrl.cache.MaterializePosition(loc)
	c.ctrl.workers.posUpdates <- posUpdateMsg{liq: l, pos: pos, loc: loc}
}

func (c *ControllerOverNetwork) IngestSwap(l tables.Swap) {
}

func (c *ControllerOverNetwork) IngestFee(l tables.FeeChange) {
}

func (c *ControllerOverNetwork) IngestKnockout(l tables.KnockoutCross) {
}

func formLiqLoc(l tables.LiqChange) types.LiquidityLocation {
	if l.PositionType == "ambient" {
		return types.AmbientLiquidityLocation()
	} else {
		return types.RangeLiquidityLocation(l.BidTick, l.AskTick)
	}
}
