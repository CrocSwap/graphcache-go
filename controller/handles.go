package controller

import (
	"log"
	"math/big"
	"time"

	"github.com/CrocSwap/graphcache-go/loader"
	"github.com/CrocSwap/graphcache-go/model"
	"github.com/CrocSwap/graphcache-go/types"
)

type IRefreshHandle interface {
	Hash() [32]byte
	RefreshTime() int64
	Skippable() bool
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
		p.pos.Liq.UpdateActiveLiq(*big.NewInt(0), time.Now().Unix())

	} else {
		claimLoc := types.KOClaimLocation{PositionLocation: p.location, PivotTime: pivotTime}
		liqFn := func() (loader.KnockoutLiqResp, error) { return (*query).QueryKnockoutLiq(claimLoc) }
		koLiqResp := tryQueryAttempt(liqFn, "knockoutLiq")
		p.pos.Liq.UpdateActiveLiq(*koLiqResp.Liq, time.Now().Unix())
	}
}

func (p *KnockoutPostHandle) RefreshQuery(query *loader.ICrocQuery) {
	liqFn := func() (loader.KnockoutLiqResp, error) { return (*query).QueryKnockoutLiq(p.location) }
	koLiqResp := tryQueryAttempt(liqFn, "knockoutLiq")
	if koLiqResp.KnockedOut {
		p.pos.Liq.UpdatePostKOLiq(p.location.PivotTime, *koLiqResp.Liq, time.Now().Unix())
	}
}

func tryQueryAttempt[T any](queryFn func() (T, error), label string) T {
	result, err := queryFn()
	for retryCount := 0; err != nil && retryCount < N_MAX_RETRIES; retryCount += 1 {
		log.Printf("Query attempt %d/%d failed for \"%s\" with err: \"%s\"", retryCount, N_MAX_RETRIES, label, err)
		retryWaitRandom()
		result, err = queryFn()
	}
	if err != nil {
		log.Fatalf("Unable to query \"%s\", err: %s", label, err)
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

func (p *PositionRefreshHandle) RefreshTime() int64 {
	return p.pos.RefreshTime
}

func (p *RewardsRefreshHandle) RefreshTime() int64 {
	return p.pos.RefreshTime
}

func (p *KnockoutAliveHandle) RefreshTime() int64 {
	return p.pos.Liq.Active.RefreshTime
}

func (p *KnockoutPostHandle) RefreshTime() int64 {
	return p.pos.Liq.Active.RefreshTime
}

func (p *PositionRefreshHandle) Hash() [32]byte {
	return p.location.Hash()
}

func (p *RewardsRefreshHandle) Hash() [32]byte {
	return p.location.Hash()
}

func (p *KnockoutAliveHandle) Hash() [32]byte {
	h := p.location.Hash()
	// Since PositionLocation for regular positions also has an IsBid bool, it's
	// possible for the hash of a knockout order to collide with a position of
	// the same user. To avoid this, we increment the first byte of the hash.
	// Sure wish Go had a non-painful way to make fields optional/nullable.
	h[0] = byte(h[0] + 1)
	return h
}

func (p *KnockoutPostHandle) Hash() [32]byte {
	return p.location.Hash()
}

func (p *PositionRefreshHandle) Skippable() bool {
	return false
}

func (p *RewardsRefreshHandle) Skippable() bool {
	return true
}

func (p *KnockoutAliveHandle) Skippable() bool {
	return false
}

func (p *KnockoutPostHandle) Skippable() bool {
	return false
}
