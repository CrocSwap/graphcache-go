package model

import (
	"github.com/CrocSwap/graphcache-go/tables"
)

type PositionTracker struct {
	PositionID       string `json:"positionId"`
	TimeFirstMint    int    `json:"timeFirstMint"`
	Time             int    `json:"time"`
	Block            int    `json:"block"`
	LatestUpdateTime int    `json:"latestUpdateTime"`
}

func (p *PositionTracker) UpdatePosition(l tables.LiqChange) {
	if p.Time == 0 {
		p.TimeFirstMint = l.Time
	}
	p.Time = l.Time
	p.Block = l.Block
	p.LatestUpdateTime = l.Time
}

type AmbientLiquidity struct {
	Seed float64 `json:"ambientSeeds"`
}

type RangeLiquidity struct {
	ConcLiq   float64 `json:"rangeLiquidity"`
	RewardLiq float64 `json:"rangeRewardLiquidity"`
}

type KnockoutLiquidity struct {
	ConcLiq       float64 `json:"rangeLiquidity"`
	RewardLiq     float64 `json:"rangeRewardLiquidity"`
	HasKnockedOut bool    `json:"hasKnockedOut"`
}
