package model

import "github.com/CrocSwap/graphcache-go/tables"

type PositionTracker struct {
	TimeFirstMint int `json:"timeFirstMint"`
	Time          int `json:"time"`
	Block         int `json:"block"`
}

func (p *PositionTracker) UpdatePosition(l tables.LiqChange) {
	if p.Time == 0 {
		p.TimeFirstMint = l.Time
	}
	p.Time = l.Time
	p.Block = l.Block
}
