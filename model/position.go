package model

import (
	"math/big"
	"time"

	"github.com/CrocSwap/graphcache-go/tables"
)

type PositionTracker struct {
	TimeFirstMint    int            `json:"timeFirstMint"`
	LatestUpdateTime int            `json:"latestUpdateTime"`
	LastMintTx       string         `json:"lastMintTx"`
	FirstMintTx      string         `json:"firstMintTx"`
	PositionType     tables.PosType `json:"positionType"`
	PositionLiquidity
	LiqHist LiquidityDeltaHist `json:"-"`
}

func (p *PositionTracker) UpdatePosition(l tables.LiqChange) {
	if p.LatestUpdateTime == 0 || l.Time < p.LatestUpdateTime {
		p.TimeFirstMint = l.Time
		if l.ChangeType == tables.ChangeTypeMint {
			p.FirstMintTx = l.TX
		}
	}
	if l.Time > p.LatestUpdateTime {
		p.LatestUpdateTime = l.Time
		if l.ChangeType == tables.ChangeTypeMint {
			p.LastMintTx = l.TX
		}
	}
	p.PositionType = l.PositionType

	p.LiqHist.appendChange(l)
}

func (p *PositionTracker) UpdateAmbient(liq big.Int) {
	p.AmbientLiq = liq
	p.RefreshTime = time.Now().Unix()
}

func (p *PositionTracker) UpdateRange(liq big.Int, rewardsLiq big.Int) {
	p.ConcLiq = liq
	p.RewardLiq = rewardsLiq
	p.RefreshTime = time.Now().Unix()
}

func (p *PositionTracker) UpdateRangeRewards(rewardsLiq big.Int) {
	p.RewardLiq = rewardsLiq
	p.RefreshTime = time.Now().Unix()
}

func (p *PositionTracker) Time() int {
	return p.LatestUpdateTime
}

func (p *PositionLiquidity) IsEmpty() bool {
	zero := big.NewInt(0)
	return p.AmbientLiq.Cmp(zero) == 0 &&
		p.ConcLiq.Cmp(zero) == 0
}

func (p *PositionLiquidity) IsConcentrated() bool {
	zero := big.NewInt(0)
	return p.ConcLiq.Cmp(zero) > 0
}

type PositionLiquidity struct {
	AmbientLiq  big.Int `json:"ambientLiq"`
	ConcLiq     big.Int `json:"concLiq"`
	RewardLiq   big.Int `json:"rewardLiq"`
	RefreshTime int64   `json:"liqRefreshTime"`
}
