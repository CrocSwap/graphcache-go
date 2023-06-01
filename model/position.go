package model

import (
	"math/big"

	"github.com/CrocSwap/graphcache-go/tables"
)

type PositionTracker struct {
	TimeFirstMint    int `json:"timeFirstMint"`
	Time             int `json:"time"`
	Block            int `json:"block"`
	LatestUpdateTime int `json:"latestUpdateTime"`
	PositionLiquidity
}

func (p *PositionTracker) UpdatePosition(l tables.LiqChange) {
	if p.Time == 0 {
		p.TimeFirstMint = l.Time
	}
	p.Time = l.Time
	p.Block = l.Block
	p.LatestUpdateTime = l.Time
}

func (p *PositionTracker) UpdateAmbient(seeds big.Int) {
	p.AmbientSeeds = seeds
}

func (p *PositionTracker) UpdateRange(liq big.Int, rewardsLiq big.Int) {
	p.ConcLiq = liq
	p.RewardLiq = rewardsLiq
}

func (p *PositionTracker) UpdateKnockout(liq big.Int, knockedOut bool) {
	p.ConcLiq = liq
	p.HasKnockedOut = knockedOut
}

type PositionLiquidity struct {
	AmbientSeeds  big.Int `json:"ambientSeeds"`
	ConcLiq       big.Int `json:"rangeLiquidity"`
	RewardLiq     big.Int `json:"rangeRewardSeeds"`
	HasKnockedOut bool    `json:"hasKnockedOut"`
}
