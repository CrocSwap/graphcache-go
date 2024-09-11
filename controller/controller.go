package controller

import (
	"log"
	"math"
	"math/big"
	"time"

	"github.com/CrocSwap/graphcache-go/cache"
	"github.com/CrocSwap/graphcache-go/loader"
	"github.com/CrocSwap/graphcache-go/model"
	"github.com/CrocSwap/graphcache-go/tables"
	"github.com/CrocSwap/graphcache-go/types"
)

type Controller struct {
	netCfg    loader.NetworkConfig
	cache     *cache.MemoryCache
	history   *model.HistoryWriter
	workers   *workers
	refresher *LiquidityRefresher
}

func New(netCfg loader.NetworkConfig, cache *cache.MemoryCache, chain *loader.OnChainLoader) *Controller {
	query := loader.NewCrocQuery(chain)

	return NewOnQuery(netCfg, cache, query)
}

func NewOnQuery(netCfg loader.NetworkConfig, cache *cache.MemoryCache, query loader.ICrocQuery) *Controller {
	history := model.NewHistoryWriter(netCfg, cache.AddPoolEvent)
	workers, refresher := initWorkers(netCfg, &query)

	ctrl := &Controller{
		netCfg:    netCfg,
		cache:     cache,
		workers:   workers,
		refresher: refresher,
		history:   history,
	}
	go ctrl.runPeriodicRefresh()

	return ctrl
}

func (c *Controller) SpinUntilLiqSync() {
	const REFRESH_PAUSE_SECS = 5
	for {
		nowTime := time.Now().Unix()
		syncSec := c.refresher.lastRefreshSec
		if nowTime < syncSec+REFRESH_PAUSE_SECS {
			// log.Println("Waiting for liquidity sync pause. Last refresh:", syncSec, "now:", nowTime)
		} else {
			return
		}
		time.Sleep(time.Second * 1)
	}
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
	if l.ChangeType == tables.ChangeTypeCross {
		c.IngestKnockout(l)
	}
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

	if l.PositionType == tables.PosTypeKnockout {
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
	pivotLoc := loc.ToBookLoc()
	if l.IsBid == 1 {
		pivotLoc.LiquidityLocation.AskTick = l.BidTick
	} else {
		pivotLoc.LiquidityLocation.BidTick = l.AskTick
	}
	if l.ChangeType == tables.ChangeTypeCross {
		c.ctrl.cache.SetPivotTime(pivotLoc, 0)
		return
	}
	pivotTime := c.ctrl.cache.RetrievePivotTime(pivotLoc)
	if l.ChangeType == tables.ChangeTypeMint && pivotTime == 0 {
		c.ctrl.cache.SetPivotTime(pivotLoc, l.Time)
		pivotTime = l.Time
	}

	event := model.KnockoutSagaTx{
		TxTime:    l.Time,
		TxHash:    l.TX,
		PivotTime: pivotTime,
	}
	pos := c.ctrl.cache.MaterializeKnockoutPos(loc)
	if l.ChangeType == tables.ChangeTypeMint {
		pos.AppendMint(event)
	} else if l.ChangeType == tables.ChangeTypeBurn {
		pos.AppendBurn(event)
	}

	c.ctrl.workers.omniUpdates <- &koPosUpdateMsg{liq: l, pos: pos, loc: loc}

	// Estimate position liquidity from the flows
	if (l.ChangeType == tables.ChangeTypeMint || l.ChangeType == tables.ChangeTypeBurn || l.ChangeType == tables.ChangeTypeRecover) && l.BaseFlow != nil && l.QuoteFlow != nil {
		liq := model.DeriveLiquidityFromConcFlow(*l.BaseFlow, *l.QuoteFlow, l.BidTick, l.AskTick)
		if math.IsInf(liq, 0) || math.IsNaN(liq) {
			log.Println("Invalid liq", liq, "for", l)
			liq = 0
		}
		liqBigInt, _ := big.NewFloat(liq).Int(nil)

		activeLiq := pos.Liq.GetActiveLiq()
		pos.Liq.UpdateActiveLiq(*big.NewInt(0).Add(activeLiq, liqBigInt), 0)
		afterLiq := pos.Liq.GetActiveLiq()
		if afterLiq.Cmp(big.NewInt(0)) < 0 {
			pos.Liq.UpdateActiveLiq(*big.NewInt(0), 0)
			afterLiq = big.NewInt(0)
		}
		afterLiqFloat, _ := afterLiq.Float64()

		// If it's a burn and the remaining liq is less than 10% of the liq change, set it to 0
		if l.ChangeType == tables.ChangeTypeBurn && afterLiqFloat > 0 && math.Abs(liq)*0.10 > math.Abs(afterLiqFloat) {
			pos.Liq.UpdateActiveLiq(*big.NewInt(0), 0)
		}
		if l.ChangeType == tables.ChangeTypeRecover || l.ChangeType == tables.ChangeTypeClaim {
			pos.Liq.UpdateActiveLiq(*big.NewInt(0), 0)
			pos.Liq.UpdatePostKOLiq(*l.PivotTime, *big.NewInt(0), 0)
		}
	}
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

func (c *ControllerOverNetwork) IngestKnockout(l tables.LiqChange) {
	liq := types.KnockoutTickLocation(l.BidTick, l.IsBid > 0, c.knockoutTickWidth())
	pool := types.PoolLocation{
		ChainId: c.chainId,
		PoolIdx: l.PoolIdx,
		Base:    types.RequireEthAddr(l.Base),
		Quote:   types.RequireEthAddr(l.Quote),
	}
	loc := types.BookLocation{
		PoolLocation:      pool,
		LiquidityLocation: liq,
	}
	pos := c.ctrl.cache.MaterializeKnockoutSaga(loc)
	pos.UpdateCross(l)
	c.ctrl.workers.omniUpdates <- &koCrossUpdateMsg{loc: loc, pos: pos, cross: l}
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
	if l.PositionType == tables.PosTypeAmbient {
		return types.AmbientLiquidityLocation()
	} else if l.PositionType == tables.PosTypeKnockout {
		return types.KnockoutRangeLocation(l.BidTick, l.AskTick, l.IsBid > 0)
	} else {
		return types.RangeLiquidityLocation(l.BidTick, l.AskTick)
	}
}

func (c *Controller) resyncFullCycle(time int) {
	for poolLoc, poolPos := range c.cache.RetrieveAllPositions() {
		if !poolPos.IsEmpty() || poolPos.RefreshTime == 0 {
			c.workers.omniUpdates <- &posImpactMsg{poolLoc, poolPos, time}
		}
	}
}

const REFRESH_CYCLE_TIME = 30 * 60

func (c *Controller) runPeriodicRefresh() {
	// To prevent running periodic refreshes while the startup sync is still running
	c.SpinUntilLiqSync()
	for {
		time.Sleep(time.Second * REFRESH_CYCLE_TIME)
		refreshTime := time.Now().Unix()
		log.Println("Running full refresh at", refreshTime)
		c.resyncFullCycle(int(refreshTime))
	}
}
