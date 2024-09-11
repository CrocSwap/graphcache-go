package controller

import (
	"math/big"

	"github.com/CrocSwap/graphcache-go/loader"
	"github.com/CrocSwap/graphcache-go/model"
	"github.com/CrocSwap/graphcache-go/tables"
	"github.com/CrocSwap/graphcache-go/types"
)

type workers struct {
	omniUpdates  chan IMsgType
	liqRefresher *LiquidityRefresher
}

func initWorkers(_ loader.NetworkConfig, query *loader.ICrocQuery) (*workers, *LiquidityRefresher) {
	liqRefresher := NewLiquidityRefresher(query)

	return &workers{
		omniUpdates:  watchUpdateSeq(liqRefresher),
		liqRefresher: NewLiquidityRefresher(query),
	}, liqRefresher
}

func watchUpdateSeq(liqRefresher *LiquidityRefresher) chan IMsgType {
	sink := make(chan IMsgType, 10000) // doesn't really need to be buffered because it gets sent to another buffer anyway

	go func() {
		for {
			msg := <-sink
			msg.processUpdate(liqRefresher)
		}
	}()
	return sink
}

type IMsgType interface {
	processUpdate(*LiquidityRefresher)
}

func (msg *posUpdateMsg) processUpdate(lr *LiquidityRefresher) {
	(msg.pos).UpdatePosition(msg.liq)
	handle := PositionRefreshHandle{location: msg.loc, pos: msg.pos}
	lr.PushRefresh(&handle, msg.liq.Time)
}

func (msg *posImpactMsg) processUpdate(lr *LiquidityRefresher) {
	handle := RewardsRefreshHandle{location: msg.loc, pos: msg.pos}
	lr.PushRefreshPoll(&handle)
}

func (msg *koPosUpdateMsg) processUpdate(lr *LiquidityRefresher) {
	cands, isPossiblyLive := (msg.pos).UpdateLiqChange(msg.liq)

	handle := KnockoutAliveHandle{location: msg.loc, pos: msg.pos}

	if isPossiblyLive {
		lr.PushRefresh(&handle, msg.liq.Time)
	}

	for _, cand := range cands {
		claimLoc := types.KOClaimLocation{PositionLocation: msg.loc, PivotTime: cand.PivotTime}
		handle := KnockoutPostHandle{location: claimLoc, pos: msg.pos}
		lr.PushRefresh(&handle, msg.liq.Time)
	}
}

func (msg *koCrossUpdateMsg) processUpdate(lr *LiquidityRefresher) {
	cands := (msg.pos).UpdateCross(msg.cross)

	for _, cand := range cands {
		claimLoc := msg.loc.ToClaimLoc(cand.User, cand.PivotTime)
		subPos := msg.pos.ForUser(cand.User)
		activeLiq := subPos.Liq.GetActiveLiq()
		subPos.Liq.UpdatePostKOLiq(cand.PivotTime, *activeLiq, 0)
		subPos.Liq.UpdateActiveLiq(*big.NewInt(0), 0)
		handle := KnockoutPostHandle{location: claimLoc, pos: subPos}
		lr.PushRefresh(&handle, msg.cross.Time)
	}
}

type posUpdateMsg struct {
	loc types.PositionLocation
	pos *model.PositionTracker
	liq tables.LiqChange
}

type posImpactMsg struct {
	loc       types.PositionLocation
	pos       *model.PositionTracker
	eventTime int
}

type koPosUpdateMsg struct {
	loc types.PositionLocation
	pos *model.KnockoutSubplot
	liq tables.LiqChange
}

type koCrossUpdateMsg struct {
	loc   types.BookLocation
	pos   *model.KnockoutSaga
	cross tables.LiqChange
}
