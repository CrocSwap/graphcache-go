package controller

import (
	"github.com/CrocSwap/graphcache-go/loader"
	"github.com/CrocSwap/graphcache-go/model"
	"github.com/CrocSwap/graphcache-go/tables"
	"github.com/CrocSwap/graphcache-go/types"
)

type workers struct {
	posUpdates     chan posUpdateMsg
	koPosUpdates   chan koPosUpdateMsg
	koCrossUpdates chan koCrossUpdateMsg
	liqRefresher   *LiquidityRefresher
	posRefreshers  map[types.PositionLocation]PositionRefresher
}

func initWorkers(netCfg loader.NetworkConfig) *workers {
	chain := &loader.OnChainLoader{Cfg: netCfg}
	query := loader.NewCrocQuery(chain)
	liqRefresher := NewLiquidityRefresher(query)

	return &workers{
		posUpdates:     watchPositionUpdates(liqRefresher),
		koPosUpdates:   watchKoPositionUpdates(liqRefresher),
		koCrossUpdates: watchKoCrossUpdates(liqRefresher),
		liqRefresher:   NewLiquidityRefresher(query),
	}
}

const UPDATE_CHANNEL_SIZE = 16000

func watchPositionUpdates(liq *LiquidityRefresher) chan posUpdateMsg {
	sink := make(chan posUpdateMsg, UPDATE_CHANNEL_SIZE)
	refreshers := make(map[types.PositionLocation]*PositionRefresher)

	go func() {
		for true {
			msg := <-sink
			(*msg.pos).UpdatePosition(msg.liq)

			refresher, ok := refreshers[msg.loc]
			if !ok {
				refresher = NewPositionRefresher(msg.loc, liq, msg.pos)
				refreshers[msg.loc] = refresher
			}

			refresher.PushRefresh(msg.liq.Time)
		}
	}()
	return sink
}

func watchKoPositionUpdates(liq *LiquidityRefresher) chan koPosUpdateMsg {
	sink := make(chan koPosUpdateMsg, UPDATE_CHANNEL_SIZE)

	go func() {
		for true {
			msg := <-sink
			(*msg.pos).UpdateLiqChange(msg.liq)
		}
	}()
	return sink
}

func watchKoCrossUpdates(liq *LiquidityRefresher) chan koCrossUpdateMsg {
	sink := make(chan koCrossUpdateMsg, UPDATE_CHANNEL_SIZE)

	go func() {
		for true {
			msg := <-sink
			(*msg.pos).UpdateCross(msg.cross)
		}
	}()
	return sink
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
