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
	chainId  types.ChainId
	chainCfg loader.ChainConfig
	ctrl     *Controller
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

	if l.PositionType == "knockout" {
		c.ingestKnockoutLiq(l, loc)
	} else {
		c.ingestPassiveLiq(l, loc)
	}
}

func (c *ControllerOverNetwork) ingestKnockoutLiq(l tables.LiqChange, loc types.PositionLocation) {
	pos := c.ctrl.cache.MaterializeKnockoutPos(loc)
	c.ctrl.workers.koPosUpdates <- koPosUpdateMsg{liq: l, pos: pos, loc: loc}
}

func (c *ControllerOverNetwork) ingestPassiveLiq(l tables.LiqChange, loc types.PositionLocation) {
	pos := c.ctrl.cache.MaterializePosition(loc)
	c.ctrl.workers.posUpdates <- posUpdateMsg{liq: l, pos: pos, loc: loc}
}

func (c *ControllerOverNetwork) IngestSwap(l tables.Swap) {
}

func (c *ControllerOverNetwork) IngestFee(l tables.FeeChange) {
}

func (c *ControllerOverNetwork) IngestKnockout(r tables.KnockoutCross) {
	liq := types.KnockoutTickLocation(r.Tick, r.IsBid > 0, c.chainCfg.KnockoutTickWidth)
	pool := types.PoolLocation{
		ChainId: c.chainId,
		PoolIdx: r.PoolIdx,
		Base:    types.RequireEthAddr(r.Base),
		Quote:   types.RequireEthAddr(r.Quote),
	}
	loc := types.BookLocation{
		PoolLocation:      pool,
		LiquidityLocation: liq,
	}
	pos := c.ctrl.cache.MaterializeKnockoutBook(loc)
	c.ctrl.workers.koCrossUpdates <- koCrossUpdateMsg{loc: loc, pos: pos, cross: r}
}

func formLiqLoc(l tables.LiqChange) types.LiquidityLocation {
	if l.PositionType == "ambient" {
		return types.AmbientLiquidityLocation()
	} else if l.PositionType == "knockout" {
		return types.KnockoutRangeLocation(l.BidTick, l.AskTick, l.IsBid > 0)
	} else {
		return types.RangeLiquidityLocation(l.BidTick, l.AskTick)
	}
}
