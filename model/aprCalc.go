package model

import (
	"math"
	"math/big"

	"github.com/CrocSwap/graphcache-go/types"
)

type APRCalcResult struct {
	Duration       float64 `json:"aprDuration"`
	PostLiqPos     float64 `json:"aprPostLiq"`
	ContributedLiq float64 `json:"aprContributedLiq"`
	Apr            float64 `json:"aprEst"`
}

func (p *PositionTracker) CalcAPR(loc types.PositionLocation) APRCalcResult {
	if p.IsEmpty() {
		return APRCalcResult{}
	}

	numerator := p.aprNumerator(loc)
	denom := p.aprDenominator()
	time := p.liqHist.weightedAverageDuration()

	apy := normalizeApr(numerator, denom, time)
	return APRCalcResult{
		Duration:       time,
		PostLiqPos:     numerator,
		ContributedLiq: denom,
		Apr:            apy,
	}
}

func (p *PositionTracker) aprNumerator(loc types.PositionLocation) float64 {
	if p.IsConcentrated() {
		amplFactor := estLiqAmplification(loc.BidTick, loc.AskTick)
		return amplFactor*castBigToFloat(&p.RewardLiq) + castBigToFloat(&p.ConcLiq)
	} else {
		return castBigToFloat(&p.AmbientLiq)
	}
}

func (p *PositionTracker) aprDenominator() float64 {
	if p.IsConcentrated() {
		return castBigToFloat(&p.ConcLiq)
	} else {
		return p.liqHist.netCumulativeLiquidity()
	}
}

const MAX_APR_CAP = 10.0

func normalizeApr(num float64, denom float64, time float64) float64 {
	growth := num / denom

	timeInYears := time / (3600 * 24 * 365)
	compounded := math.Pow(growth, 1.0/timeInYears) - 1.0

	if compounded < 0.0 {
		return 0.0
	} else if compounded > MAX_APR_CAP {
		return MAX_APR_CAP
	}
	return compounded
}

func castBigToFloat(liq *big.Int) float64 {
	f := new(big.Float).SetInt(liq)
	ret, _ := f.Float64()
	return ret
}
