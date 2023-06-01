package controller

import (
	"github.com/CrocSwap/graphcache-go/loader"
	"github.com/CrocSwap/graphcache-go/model"
	"github.com/CrocSwap/graphcache-go/tables"
	"github.com/CrocSwap/graphcache-go/types"
)

type workers struct {
	posUpdates    chan posUpdateMsg
	liqRefresher  *LiquidityRefresher
	posRefreshers map[types.PositionLocation]PositionRefresher
}

func initWorkers(netCfg loader.NetworkConfig) *workers {
	chain := &loader.OnChainLoader{Cfg: netCfg}
	query := loader.NewCrocQuery(chain)
	liqRefresher := NewLiquidityRefresher(query)

	return &workers{
		posUpdates:   watchPositionUpdates(liqRefresher),
		liqRefresher: NewLiquidityRefresher(query),
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

			refresher.PushRefresh()
		}
	}()
	return sink
}

type posUpdateMsg struct {
	loc types.PositionLocation
	pos *model.PositionTracker
	liq tables.LiqChange
}
