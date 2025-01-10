package controller

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"log"
	"math/big"
	"time"

	"github.com/CrocSwap/graphcache-go/loader"
	"github.com/CrocSwap/graphcache-go/model"
	"github.com/CrocSwap/graphcache-go/types"
)

type IRefreshHandle interface {
	Hash(buf *bytes.Buffer) [32]byte
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

type PoolInitPriceHandle struct {
	Pool  types.PoolLocation
	Block int
	Hist  *model.PoolTradingHistory
}

type BumpRefreshHandle struct {
	pool  types.PoolLocation
	tick  int
	curve *model.LiquidityCurve
	bump  *model.LiquidityBump
}

func (p *PositionRefreshHandle) RefreshQuery(query *loader.ICrocQuery) {
	posType := types.PositionTypeForLiq(p.location.LiquidityLocation)

	if posType == "ambient" {
		liqFn := func() (*big.Int, error) { return (*query).QueryAmbientLiq(p.location) }
		ambientLiq, _ := tryQueryAttempt(liqFn, "ambientLiq", N_MAX_RETRIES, true)
		p.pos.UpdateAmbient(*ambientLiq)
	}

	if posType == "range" {
		liqFn := func() (*big.Int, error) { return (*query).QueryRangeLiquidity(p.location) }
		rewardFn := func() (*big.Int, error) { return (*query).QueryRangeRewardsLiq(p.location) }
		concLiq, _ := tryQueryAttempt(liqFn, "rangeLiq", N_MAX_RETRIES, true)
		rewardLiq, _ := tryQueryAttempt(rewardFn, "rangeRewards", N_MAX_RETRIES, true)
		p.pos.UpdateRange(*concLiq, *rewardLiq)
	}
}

func (p *RewardsRefreshHandle) RefreshQuery(query *loader.ICrocQuery) {
	posType := types.PositionTypeForLiq(p.location.LiquidityLocation)

	if posType == "range" {
		rewardFn := func() (*big.Int, error) { return (*query).QueryRangeRewardsLiq(p.location) }
		rewardLiq, _ := tryQueryAttempt(rewardFn, "rangeRewards", N_MAX_RETRIES, true)
		p.pos.UpdateRangeRewards(*rewardLiq)
	}
}

func (p *KnockoutAliveHandle) RefreshQuery(query *loader.ICrocQuery) {
	pivotTimeFn := func() (uint32, error) { return (*query).QueryKnockoutPivot(p.location) }
	pivotTime, _ := tryQueryAttempt(pivotTimeFn, "pivotTimeLatest", N_MAX_RETRIES, true)

	if pivotTime == 0 {
		p.pos.Liq.UpdateActiveLiq(*big.NewInt(0), time.Now().Unix())

	} else {
		claimLoc := types.KOClaimLocation{PositionLocation: p.location, PivotTime: int(pivotTime)}
		liqFn := func() (loader.KnockoutLiqResp, error) { return (*query).QueryKnockoutLiq(claimLoc) }
		koLiqResp, _ := tryQueryAttempt(liqFn, "knockoutLiq", N_MAX_RETRIES, true)
		p.pos.Liq.UpdateActiveLiq(*koLiqResp.Liq, time.Now().Unix())
	}
}

func (p *KnockoutPostHandle) RefreshQuery(query *loader.ICrocQuery) {
	liqFn := func() (loader.KnockoutLiqResp, error) { return (*query).QueryKnockoutLiq(p.location) }
	koLiqResp, _ := tryQueryAttempt(liqFn, "knockoutLiq", N_MAX_RETRIES, true)
	if koLiqResp.KnockedOut {
		p.pos.Liq.UpdatePostKOLiq(p.location.PivotTime, *koLiqResp.Liq, time.Now().Unix())
	}
}

func (p *PoolInitPriceHandle) RefreshQuery(query *loader.ICrocQuery) {
	// priceFn := func() (*big.Int, error) { return (*query).QueryPoolPrice(p.Pool) }
}

func (p *BumpRefreshHandle) RefreshQuery(query *loader.ICrocQuery) {
	refreshFn := func() (loader.LevelResp, error) { return (*query).QueryLevel(p.pool, p.tick) }
	levelResp, _ := tryQueryAttempt(refreshFn, "bumpRefresh", N_MAX_RETRIES, false)
	askLiq := big.NewInt(0).Mul(levelResp.AskLots, big.NewInt(1024))
	bidLiq := big.NewInt(0).Mul(levelResp.BidLots, big.NewInt(1024))
	delta := big.NewInt(0).Sub(bidLiq, askLiq)
	deltaF64, _ := delta.Float64()
	p.bump.LiquidityDelta = deltaF64
}

func tryQueryAttempt[T any](queryFn func() (T, error), label string, nAttempts int, fatal bool) (result T, err error) {
	result, err = queryFn()
	for retryCount := 0; err != nil && retryCount < nAttempts; retryCount += 1 {
		log.Printf("Query attempt %d/%d failed for \"%s\" with err: \"%s\"", retryCount, nAttempts, label, err)
		retryWaitRandom()
		result, err = queryFn()
	}
	if err != nil {
		if fatal {
			log.Fatalf("Unable to query \"%s\", err: %s", label, err)
		} else {
			log.Printf("Unable to query \"%s\", err: %s, giving up", label, err)
		}
	}
	return
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

func (p *PoolInitPriceHandle) LabelTag() string {
	return "poolInitPrice"
}

func (p *BumpRefreshHandle) LabelTag() string {
	return "bumpRefresh"
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

func (p *PoolInitPriceHandle) RefreshTime() int64 {
	return 0
}

func (p *BumpRefreshHandle) RefreshTime() int64 {
	return 0
}

func (p *PositionRefreshHandle) Hash(buf *bytes.Buffer) [32]byte {
	return p.location.Hash(buf)
}

func (p *RewardsRefreshHandle) Hash(buf *bytes.Buffer) [32]byte {
	return p.location.Hash(buf)
}

func (p *KnockoutAliveHandle) Hash(buf *bytes.Buffer) [32]byte {
	h := p.location.Hash(buf)
	// Since PositionLocation for regular positions also has an IsBid bool, it's
	// possible for the hash of a knockout order to collide with a position of
	// the same user. To avoid this, we increment the first byte of the hash.
	// Sure wish Go had a non-painful way to make fields optional/nullable.
	h[0] = byte(h[0] + 1)
	return h
}

func (p *KnockoutPostHandle) Hash(buf *bytes.Buffer) [32]byte {
	return p.location.Hash(buf)
}

func (p *PoolInitPriceHandle) Hash(buf *bytes.Buffer) [32]byte {
	return p.Pool.Hash(buf)
}

func (p *BumpRefreshHandle) Hash(buf *bytes.Buffer) [32]byte {
	if buf == nil {
		buf = new(bytes.Buffer)
		buf.Grow(100)
	} else {
		buf.Reset()
	}
	buf.WriteString(string(p.pool.ChainId))
	buf.WriteString(string(p.pool.Base))
	buf.WriteString(string(p.pool.Quote))
	binary.Write(buf, binary.BigEndian, int32(p.pool.PoolIdx))
	binary.Write(buf, binary.BigEndian, int32(p.tick))
	return sha256.Sum256(buf.Bytes())
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

func (p *PoolInitPriceHandle) Skippable() bool {
	return false
}

func (p *BumpRefreshHandle) Skippable() bool {
	return false
}
