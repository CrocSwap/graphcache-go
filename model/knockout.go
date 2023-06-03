package model

import (
	"log"
	"math/big"

	"github.com/CrocSwap/graphcache-go/tables"
	"github.com/CrocSwap/graphcache-go/types"
)

type KnockoutSubplot struct {
	mints []KnockoutSagaTx
	burns []KnockoutSagaTx
	saga  *KnockoutSaga
	Liq   KnockoutLiquiditySeries
}

type KnockoutLiquiditySeries struct {
	active     PositionLiquidity
	knockedOut map[int]*KnockedOutPostLiq
}

type KnockedOutPostLiq struct {
	PositionLiquidity
	CrossTime int `json:"crossTime"`
}

type KnockoutSaga struct {
	users   map[types.EthAddress]*KnockoutSubplot
	crosses []KnockoutSagaCross
}

type KnockoutSagaTx struct {
	TxTime int
	TxHash string
}

type KnockoutSagaCross struct {
	CrossTime int
	PivotTime int
}

type KnockoutPivotCands struct {
	PivotTime int
	User      types.EthAddress
}

func NewKnockoutSaga() *KnockoutSaga {
	return &KnockoutSaga{
		users:   make(map[types.EthAddress]*KnockoutSubplot),
		crosses: make([]KnockoutSagaCross, 0),
	}
}

func (k *KnockoutSaga) ForUser(user types.EthAddress) *KnockoutSubplot {
	subplot, ok := k.users[user]
	if !ok {
		liq := KnockoutLiquiditySeries{
			knockedOut: make(map[int]*KnockedOutPostLiq, 0),
		}
		subplot = &KnockoutSubplot{
			mints: make([]KnockoutSagaTx, 0),
			burns: make([]KnockoutSagaTx, 0),
			saga:  k,
			Liq:   liq,
		}
		k.users[user] = subplot
	}
	return subplot
}

func (k *KnockoutSubplot) UpdateLiqChange(l tables.LiqChange) []KnockoutPivotCands {
	event := KnockoutSagaTx{
		TxTime: l.Time,
		TxHash: l.TX,
	}

	if l.ChangeType == "mint" {
		k.mints = append(k.mints, event)
		return k.scrapePivotsCandsOnMint(l.Time, types.RequireEthAddr(l.User))

	} else if l.ChangeType == "burn" {
		k.burns = append(k.burns, event)
		return make([]KnockoutPivotCands, 0)

	} else if l.PivotTime != nil && *l.PivotTime > 0 {
		cand := KnockoutPivotCands{
			PivotTime: *l.PivotTime,
			User:      types.EthAddress(l.User),
		}
		return []KnockoutPivotCands{cand}

	} else {
		log.Println("Warning: Missing pivot time on knockout liq change that's not a mint or burn")
		return make([]KnockoutPivotCands, 0)
	}
}

func (k *KnockoutSaga) UpdateCross(l tables.KnockoutCross) []KnockoutPivotCands {
	event := KnockoutSagaCross{
		CrossTime: l.Time,
		PivotTime: l.PivotTime,
	}
	k.crosses = append(k.crosses, event)
	return k.scrapePivotsCandsOnCross(l.PivotTime, l.Time)
}

func (k *KnockoutSubplot) scrapePivotsCandsOnMint(mintTime int, user types.EthAddress) []KnockoutPivotCands {
	cands := make([]KnockoutPivotCands, 0)
	for _, cross := range (*k.saga).crosses {
		if isMintMaybeInPiovt(mintTime, cross.PivotTime, cross.CrossTime) {
			cands = append(cands, KnockoutPivotCands{
				PivotTime: cross.PivotTime,
				User:      user,
			})
		}
	}
	return cands
}

func (k *KnockoutSaga) scrapePivotsCandsOnCross(pivotTime int, crossTime int) []KnockoutPivotCands {
	cands := make([]KnockoutPivotCands, 0)
	for userAddr, subplot := range k.users {
		for _, mint := range subplot.mints {
			if isMintMaybeInPiovt(mint.TxTime, pivotTime, crossTime) {
				cands = append(cands, KnockoutPivotCands{
					PivotTime: pivotTime,
					User:      userAddr,
				})
			}
		}
	}
	return cands
}

/* Returns true if there's a possibility that the minted liquidity may be knocked out on the
 * the pivot cross event. */
func isMintMaybeInPiovt(mintTime int, pivotTime int, knockoutTime int) bool {
	return mintTime >= pivotTime && mintTime <= knockoutTime
}

func (k *KnockoutLiquiditySeries) UpdateActiveLiq(liqQty big.Int) {
	k.active.ConcLiq = liqQty
}

func (k *KnockoutLiquiditySeries) UpdatePostKOLiq(pivotTime int, liqQty big.Int) {
	posLiq, ok := k.knockedOut[pivotTime]
	if !ok {
		posLiq = &KnockedOutPostLiq{}
		k.knockedOut[pivotTime] = &KnockedOutPostLiq{}
	}
	posLiq.ConcLiq = liqQty
}

func (k *KnockoutLiquiditySeries) UpdateKOCrossLiq(pivotTime int, crossTime int) {
	posLiq, ok := k.knockedOut[pivotTime]
	if !ok {
		posLiq = &KnockedOutPostLiq{}
		k.knockedOut[pivotTime] = &KnockedOutPostLiq{}
	}
	posLiq.CrossTime = crossTime
}
