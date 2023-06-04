package model

import (
	"math/big"

	"github.com/CrocSwap/graphcache-go/tables"
)

type PositionTracker struct {
	TimeFirstMint    int    `json:"timeFirstMint"`
	LatestUpdateTime int    `json:"latestUpdateTime"`
	LastMintTx       string `json:"lastMintTx"`
	FirstMintTx      string `json:"firstMintTx"`
	PositionType     string `json:"positionType"`
	PositionLiquidity
	liqHist LiquidityDeltaHist
}

func (p *PositionTracker) UpdatePosition(l tables.LiqChange) {
	if p.LatestUpdateTime == 0 || l.Time < p.LatestUpdateTime {
		p.TimeFirstMint = l.Time
		if l.ChangeType == "mint" {
			p.FirstMintTx = l.TX
		}
	}
	if l.Time > p.LatestUpdateTime {
		p.LatestUpdateTime = l.Time
		if l.ChangeType == "mint" {
			p.LastMintTx = l.TX
		}
	}
	p.PositionType = l.PositionType

	p.liqHist.appendChange(l)
}

func (p *PositionTracker) UpdateAmbient(seeds big.Int) {
	p.AmbientSeeds = seeds
}

func (p *PositionTracker) UpdateRange(liq big.Int, rewardsLiq big.Int) {
	p.ConcLiq = liq
	p.RewardLiq = rewardsLiq
}

func (p *PositionLiquidity) IsEmpty() bool {
	zero := big.NewInt(0)
	return p.AmbientSeeds.Cmp(zero) == 0 &&
		p.ConcLiq.Cmp(zero) == 0
}

type PositionLiquidity struct {
	AmbientSeeds big.Int `json:"ambientSeeds"`
	ConcLiq      big.Int `json:"concLiq"`
	RewardLiq    big.Int `json:"rewardLiq"`
}
