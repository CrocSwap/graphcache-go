package controller

import (
	"log"
	"math/big"

	"github.com/CrocSwap/graphcache-go/loader"
	"github.com/CrocSwap/graphcache-go/model"
	"github.com/CrocSwap/graphcache-go/types"
)

type IRefreshHandle interface {
	RefreshQuery(query *loader.ICrocQuery)
	LabelTag() string
}

type PositionRefreshHandle struct {
	location types.PositionLocation
	pos      *model.PositionTracker
}
type RewardsRefreshHandle struct {
	location types.PositionLocation
	pos      *model.PositionTracker
}

type KnockoutAliveHandle struct {
	location types.PositionLocation
	pos      *model.KnockoutSubplot
}

type KnockoutPostHandle struct {
	location types.KOClaimLocation
	pos      *model.KnockoutSubplot
}

func (p *PositionRefreshHandle) RefreshQuery(query *loader.ICrocQuery) {
	posType := types.PositionTypeForLiq(p.location.LiquidityLocation)

	if posType == "ambient" {
		liqFn := func() (*big.Int, error) { return (*query).QueryAmbientLiq(p.location) }
		ambientLiq := tryQueryAttempt(liqFn, "ambientLiq")
		p.pos.UpdateAmbient(*ambientLiq)
	}

	if posType == "range" {
		liqFn := func() (*big.Int, error) { return (*query).QueryRangeLiquidity(p.location) }
		rewardFn := func() (*big.Int, error) { return (*query).QueryRangeRewardsLiq(p.location) }
		concLiq := tryQueryAttempt(liqFn, "rangeLiq")
		rewardLiq := tryQueryAttempt(rewardFn, "rangeRewards")
		p.pos.UpdateRange(*concLiq, *rewardLiq)
	}
}

func (p *RewardsRefreshHandle) RefreshQuery(query *loader.ICrocQuery) {
	posType := types.PositionTypeForLiq(p.location.LiquidityLocation)

	if posType == "range" {
		rewardFn := func() (*big.Int, error) { return (*query).QueryRangeRewardsLiq(p.location) }
		rewardLiq := tryQueryAttempt(rewardFn, "rangeRewards")
		p.pos.UpdateRangeRewards(*rewardLiq)
	}
}

func (p *KnockoutAliveHandle) RefreshQuery(query *loader.ICrocQuery) {
	pivotTimeFn := func() (uint32, error) { return (*query).QueryKnockoutPivot(p.location) }
	pivotTime := int(tryQueryAttempt(pivotTimeFn, "pivotTimeLatest"))

	if pivotTime == 0 {
		p.pos.Liq.UpdateActiveLiq(*big.NewInt(0))

	} else {
		claimLoc := types.KOClaimLocation{PositionLocation: p.location, PivotTime: pivotTime}
		liqFn := func() (*big.Int, error) { return (*query).QueryKnockoutLiq(claimLoc) }

		knockoutLiq := tryQueryAttempt(liqFn, "knockoutLiq")
		p.pos.Liq.UpdateActiveLiq(*knockoutLiq)
	}
}

func (p *KnockoutPostHandle) RefreshQuery(query *loader.ICrocQuery) {
	liqFn := func() (*big.Int, error) { return (*query).QueryKnockoutLiq(p.location) }
	knockoutLiq := tryQueryAttempt(liqFn, "knockoutLiq")
	p.pos.Liq.UpdatePostKOLiq(p.location.PivotTime, *knockoutLiq)
}

func tryQueryAttempt[T any](queryFn func() (T, error), label string) T {
	result, err := queryFn()
	for retryCount := 0; err != nil && retryCount < N_MAX_RETRIES; retryCount += 1 {
		retryWaitRandom()
		result, err = queryFn()
	}
	if err != nil {
		log.Fatal("Unable to query liquidity for " + label)
	}
	return result
}

func (p *PositionRefreshHandle) LabelTag() string {
	return types.PositionTypeForLiq(p.location.LiquidityLocation)
}

func (p *RewardsRefreshHandle) LabelTag() string {
	return "rewards-" + types.PositionTypeForLiq(p.location.LiquidityLocation)
}

func (p *KnockoutAliveHandle) LabelTag() string {
	return "knockoutActive"
}

func (p *KnockoutPostHandle) LabelTag() string {
	return "knockoutPost"
}
