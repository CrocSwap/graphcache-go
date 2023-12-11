package controller

import (
	"log"
	"time"

	"github.com/CrocSwap/graphcache-go/cache"
	"github.com/CrocSwap/graphcache-go/loader"
	"github.com/CrocSwap/graphcache-go/model"
	"github.com/CrocSwap/graphcache-go/tables"
	"github.com/CrocSwap/graphcache-go/types"
)

type Controller struct {
	netCfg  loader.NetworkConfig
	cache   *cache.MemoryCache
	history *model.HistoryWriter
	workers *workers
}

func New(netCfg loader.NetworkConfig, cache *cache.MemoryCache) *Controller {
	chain := &loader.OnChainLoader{Cfg: netCfg}
	query := loader.NewCrocQuery(chain)

	return NewOnQuery(netCfg, cache, query)
}

func NewOnQuery(netCfg loader.NetworkConfig, cache *cache.MemoryCache, query loader.ICrocQuery) *Controller {
	history := model.NewHistoryWriter(netCfg, cache.AddPoolEvent)

	ctrl := &Controller{
		netCfg:  netCfg,
		cache:   cache,
		workers: initWorkers(netCfg, &query),
		history: history,
	}
	go ctrl.runPeriodicRefresh()

	return ctrl
}

type ControllerOverNetwork struct {
	chainId  types.ChainId
	chainCfg loader.ChainConfig
	ctrl     *Controller
}

func (c *Controller) OnNetwork(network types.NetworkName) *ControllerOverNetwork {
	chainId, okay := c.netCfg.ChainIDForNetwork(network)
	chainCfg, okay2 := c.netCfg.ChainConfig(chainId)
	if !okay || !okay2 {
		log.Fatal("No network config for " + network)
	}

	return &ControllerOverNetwork{
		ctrl:     c,
		chainId:  chainId,
		chainCfg: chainCfg,
	}
}

func (c *ControllerOverNetwork) IngestBalance(b tables.Balance) {
	token := types.RequireEthAddr(b.Token)
	user := types.RequireEthAddr(b.User)
	c.ctrl.cache.AddUserBalance(c.chainId, user, token)
}

func (c *ControllerOverNetwork) IngestLiqChange(l tables.LiqChange) {
	c.applyToPosition(l)
	c.applyToLiqCurve(l)
	c.ctrl.history.CommitLiqChange(l)
}

func (c *ControllerOverNetwork) applyToPosition(l tables.LiqChange) {
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
		c.applyToKnockout(l, loc)
	} else {
		c.applyToPassiveLiq(l, loc)
	}

}

func (c *ControllerOverNetwork) applyToLiqCurve(l tables.LiqChange) {
	pool := types.PoolLocation{
		ChainId: c.chainId,
		PoolIdx: l.PoolIdx,
		Base:    types.RequireEthAddr(l.Base),
		Quote:   types.RequireEthAddr(l.Quote),
	}
	curve := c.ctrl.cache.MaterializePoolLiqCurve(pool)
	curve.UpdateLiqChange(l)
}

func (c *ControllerOverNetwork) applyToKnockout(l tables.LiqChange, loc types.PositionLocation) {
	if l.ChangeType == "cross" {
		return // Cross events are handled KnockoutCross table
	}
	pos := c.ctrl.cache.MaterializeKnockoutPos(loc)
	c.ctrl.workers.omniUpdates <- &koPosUpdateMsg{liq: l, pos: pos, loc: loc}
}

func (c *ControllerOverNetwork) applyToPassiveLiq(l tables.LiqChange, loc types.PositionLocation) {
	pos := c.ctrl.cache.MaterializePosition(loc)
	c.ctrl.workers.omniUpdates <- &posUpdateMsg{liq: l, pos: pos, loc: loc}
}

func (c *ControllerOverNetwork) IngestSwap(l tables.Swap) {
	c.ctrl.history.CommitSwap(l)

	updates := c.resyncPoolOnSwap(l)
	// Use array entry, instead of element loop, because otherwise same pointer
	// is passed multiple times to channel and may overwritten
	for i := range updates {
		c.ctrl.workers.omniUpdates <- &updates[i]
	}
}

func (c *ControllerOverNetwork) IngestFee(l tables.FeeChange) {
}

func (c *ControllerOverNetwork) IngestAggEvent(r tables.AggEvent) {
	pool := types.PoolLocation{
		ChainId: c.chainId,
		PoolIdx: r.PoolIdx,
		Base:    types.RequireEthAddr(r.Base),
		Quote:   types.RequireEthAddr(r.Quote),
	}
	hist := c.ctrl.cache.MaterializePoolTradingHist(pool)
	hist.NextEvent(r)
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
	c.ctrl.workers.omniUpdates <- &koCrossUpdateMsg{loc: loc, pos: pos, cross: r}
}

/* Called to indicate that all tables have completed the most recent sync cycle up
 * to the checkpointed time. */
func (c *ControllerOverNetwork) FlushSyncCycle(time int) {

}

/* Currently this uses a preset value from the network config. Long-term we should be querying
 * the knockout position from CrocQuery::queryKnockoutPivot() to confirm tick widthm, as it
 * can change over time, and different tick widths could even co-exist in the same pool. */
func (c *ControllerOverNetwork) knockoutTickWidth() int {
	return c.chainCfg.KnockoutTickWidth
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

func (c *Controller) resyncFullCycle(time int) {
	for poolLoc, poolPos := range c.cache.RetrieveAllPositions() {
		if !poolPos.IsEmpty() {
			c.workers.omniUpdates <- &posImpactMsg{poolLoc, poolPos, time}
		}
	}
}

const REFRESH_CYCLE_TIME = 30 * 60

func (c *Controller) runPeriodicRefresh() {
	for true {
		time.Sleep(time.Second * REFRESH_CYCLE_TIME)
		refreshTime := time.Now().Unix()
		c.resyncFullCycle(int(refreshTime))
	}
}
