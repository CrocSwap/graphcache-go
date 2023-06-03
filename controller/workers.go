package controller

import (
	"github.com/CrocSwap/graphcache-go/loader"
	"github.com/CrocSwap/graphcache-go/model"
	"github.com/CrocSwap/graphcache-go/tables"
	"github.com/CrocSwap/graphcache-go/types"
)

type workers struct {
	omniUpdates  chan IMsgType
	liqRefresher *LiquidityRefresher
}

func initWorkers(netCfg loader.NetworkConfig) *workers {
	chain := &loader.OnChainLoader{Cfg: netCfg}
	query := loader.NewCrocQuery(chain)
	liqRefresher := NewLiquidityRefresher(query)

	return &workers{
		omniUpdates:  watchUpdateSeq(liqRefresher),
		liqRefresher: NewLiquidityRefresher(query),
	}
}

type RefreshAccumulator struct {
	posRefreshers    map[types.PositionLocation]*HandleRefresher
	koLiveRefreshers map[types.PositionLocation]*HandleRefresher
	koPostRefreshers map[types.KOClaimLocation]*HandleRefresher
}

func watchUpdateSeq(liq *LiquidityRefresher) chan IMsgType {
	sink := make(chan IMsgType, UPDATE_CHANNEL_SIZE)

	accum := &RefreshAccumulator{
		posRefreshers:    make(map[types.PositionLocation]*HandleRefresher),
		koLiveRefreshers: make(map[types.PositionLocation]*HandleRefresher),
		koPostRefreshers: make(map[types.KOClaimLocation]*HandleRefresher),
	}

	go func() {
		for true {
			msg := <-sink
			msg.processUpdate(accum, liq)
		}
	}()
	return sink
}

const UPDATE_CHANNEL_SIZE = 16000

type IMsgType interface {
	processUpdate(*RefreshAccumulator, *LiquidityRefresher)
}

func (msg *posUpdateMsg) processUpdate(accum *RefreshAccumulator, liq *LiquidityRefresher) {
	(msg.pos).UpdatePosition(msg.liq)

	refresher, ok := accum.posRefreshers[msg.loc]
	if !ok {
		handle := PositionRefreshHandle{location: msg.loc, pos: msg.pos}
		refresher = NewHandleRefresher(&handle, liq.pending)
		accum.posRefreshers[msg.loc] = refresher
	}
	refresher.PushRefresh(msg.liq.Time)
}

func (msg *koPosUpdateMsg) processUpdate(accum *RefreshAccumulator, liq *LiquidityRefresher) {
	cands, isPossiblyLive := (msg.pos).UpdateLiqChange(msg.liq)

	refresher, ok := accum.koLiveRefreshers[msg.loc]
	if !ok {
		handle := KnockoutAliveHandle{location: msg.loc, pos: msg.pos}
		refresher = NewHandleRefresher(&handle, liq.pending)
		accum.koLiveRefreshers[msg.loc] = refresher
	}

	if isPossiblyLive {
		refresher.PushRefresh(msg.liq.Time)
	}

	for _, cand := range cands {
		claimLoc := types.KOClaimLocation{PositionLocation: msg.loc, PivotTime: cand.PivotTime}
		refresher, ok := accum.koPostRefreshers[claimLoc]
		if !ok {
			handle := KnockoutPostHandle{location: claimLoc, pos: msg.pos}
			refresher = NewHandleRefresher(&handle, liq.pending)
			accum.koPostRefreshers[claimLoc] = refresher
		}
		refresher.PushRefresh(msg.liq.Time)
	}
}

func (msg *koCrossUpdateMsg) processUpdate(accum *RefreshAccumulator, liq *LiquidityRefresher) {
	cands := (msg.pos).UpdateCross(msg.cross)

	for _, cand := range cands {
		claimLoc := msg.loc.ToClaimLoc(cand.User, cand.PivotTime)
		refresher, ok := accum.koPostRefreshers[claimLoc]

		if !ok {
			subPos := msg.pos.ForUser(cand.User)
			handle := KnockoutPostHandle{location: claimLoc, pos: subPos}
			refresher = NewHandleRefresher(&handle, liq.pending)
			accum.koPostRefreshers[claimLoc] = refresher
		}
		refresher.PushRefresh(msg.cross.Time)
	}
}

type posUpdateMsg struct {
	loc types.PositionLocation
	pos *model.PositionTracker
	liq tables.LiqChange
}

type koPosUpdateMsg struct {
	loc types.PositionLocation
	pos *model.KnockoutSubplot
	liq tables.LiqChange
}

type koCrossUpdateMsg struct {
	loc   types.BookLocation
	pos   *model.KnockoutSaga
	cross tables.KnockoutCross
}
