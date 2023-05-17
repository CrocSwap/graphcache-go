package controller

import (
	"github.com/CrocSwap/graphcache-go/model"
	"github.com/CrocSwap/graphcache-go/tables"
)

type workers struct {
	posUpdates chan posUpdateMsg
}

func initWorkers() *workers {
	return &workers{
		posUpdates: watchPositionUpdates(),
	}
}

const UPDATE_CHANNEL_SIZE = 16000

func watchPositionUpdates() chan posUpdateMsg {
	sink := make(chan posUpdateMsg, UPDATE_CHANNEL_SIZE)
	go func() {
		for true {
			msg := <-sink
			(*msg.pos).UpdatePosition(msg.liq)
		}
	}()
	return sink
}

type posUpdateMsg struct {
	pos *model.PositionTracker
	liq tables.LiqChange
}
