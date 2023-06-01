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
}

func (p *PositionTracker) UpdatePosition(l tables.LiqChange) {
	if p.LatestUpdateTime == 0 {
		p.TimeFirstMint = l.Time
		p.FirstMintTx = l.TX
	}
	if l.ChangeType == "mint" {
		p.LastMintTx = l.TX
	}
	p.PositionType = l.PositionType
	p.LatestUpdateTime = l.Time
}

func (p *PositionTracker) UpdateAmbient(seeds big.Int) {
	p.AmbientSeeds = seeds
}

func (p *PositionTracker) UpdateRange(liq big.Int, rewardsLiq big.Int) {
	p.ConcLiq = liq
	p.RewardLiq = rewardsLiq
}

type PositionLiquidity struct {
	AmbientSeeds big.Int `json:"ambientSeeds"`
	ConcLiq      big.Int `json:"concLiq"`
	RewardLiq    big.Int `json:"rewardLiq"`
}
