package model

import (
	"log"
	"math/big"
	"slices"
	"sync"

	"github.com/CrocSwap/graphcache-go/tables"
	"github.com/CrocSwap/graphcache-go/types"
)

type KnockoutSubplot struct {
	Mints            []KnockoutSagaTx
	Burns            []KnockoutSagaTx
	saga             *KnockoutSaga
	Liq              KnockoutLiquiditySeries
	LatestUpdateTime int

	// Lock is needed here (and in KnockoutLiquiditySeries) because of data
	// races between the liquidity refresher and ingestion workers. Unfortunate,
	// but
	lock sync.Mutex
}

type KnockoutLiquiditySeries struct {
	Active        PositionLiquidity
	KnockedOut    map[int]*PositionLiquidity
	TimeFirstMint int
	lock          sync.Mutex
}

type KnockoutSaga struct {
	users   map[types.EthAddress]*KnockoutSubplot
	crosses []KnockoutSagaCross
	lock    sync.Mutex
}

type KnockoutSagaTx struct {
	TxTime    int
	TxHash    string
	PivotTime int
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
		lock:    sync.Mutex{},
	}
}

func (k *KnockoutSubplot) IsActiveEmpty() bool {
	zero := big.NewInt(0)
	return k.Liq.Active.AmbientLiq.Cmp(zero) == 0
}

func (k *KnockoutSubplot) GetCrossForPivotTime(pivotTime int) (int, bool) {
	for _, cross := range k.saga.crosses {
		if cross.PivotTime == pivotTime {
			return cross.CrossTime, true
		}
	}
	return -1, false
}

func (k *KnockoutSaga) ForUser(user types.EthAddress) *KnockoutSubplot {
	k.lock.Lock()
	defer k.lock.Unlock()
	subplot, ok := k.users[user]
	if !ok {
		subplot = &KnockoutSubplot{
			Mints: make([]KnockoutSagaTx, 0),
			Burns: make([]KnockoutSagaTx, 0),
			saga:  k,
			Liq: KnockoutLiquiditySeries{
				KnockedOut: make(map[int]*PositionLiquidity, 0),
				lock:       sync.Mutex{},
			},
			lock: sync.Mutex{},
		}
		k.users[user] = subplot
	}
	return subplot
}

func (k *KnockoutSubplot) UpdateLiqChange(l tables.LiqChange) ([]KnockoutPivotCands, bool) {
	k.lock.Lock()
	defer k.lock.Unlock()
	if l.Time > k.LatestUpdateTime {
		k.LatestUpdateTime = l.Time
	}

	// By definition, only mint can occur first, so no need to check to see if chnage is a mint
	if k.Liq.TimeFirstMint == 0 || l.Time < k.Liq.TimeFirstMint {
		k.Liq.TimeFirstMint = l.Time
	}

	if l.ChangeType == tables.ChangeTypeMint {
		return k.scrapePivotsCandsOnMint(l.Time, types.RequireEthAddr(l.User)), true

	} else if l.ChangeType == tables.ChangeTypeBurn {
		return make([]KnockoutPivotCands, 0), true

	} else if l.PivotTime != nil && *l.PivotTime > 0 {
		cand := KnockoutPivotCands{
			PivotTime: *l.PivotTime,
			User:      types.EthAddress(l.User),
		}
		return []KnockoutPivotCands{cand}, false

	} else {
		log.Println("Warning: Missing pivot time on knockout liq change that's not a mint or burn")
		return make([]KnockoutPivotCands, 0), false
	}
}

func (k *KnockoutSubplot) AppendMint(mint KnockoutSagaTx) {
	k.lock.Lock()
	defer k.lock.Unlock()
	k.Mints = append(k.Mints, mint)
}

func (k *KnockoutSubplot) AppendBurn(burn KnockoutSagaTx) {
	k.lock.Lock()
	defer k.lock.Unlock()
	k.Burns = append(k.Burns, burn)
}

func (k *KnockoutSubplot) Time() int {
	return k.LatestUpdateTime
}

func (k *KnockoutSaga) UpdateCross(l tables.LiqChange) []KnockoutPivotCands {
	k.lock.Lock()
	defer k.lock.Unlock()
	event := KnockoutSagaCross{
		CrossTime: l.Time,
		PivotTime: *l.PivotTime,
	}
	k.crosses = append(k.crosses, event)
	return k.scrapePivotsCandsOnCross(*l.PivotTime, l.Time)
}

func (k *KnockoutSubplot) scrapePivotsCandsOnMint(mintTime int, user types.EthAddress) []KnockoutPivotCands {
	cands := make([]KnockoutPivotCands, 0)
	for _, cross := range (*k.saga).crosses {
		if isMintMaybeInPivot(mintTime, cross.PivotTime, cross.CrossTime) {
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
		for _, mint := range subplot.Mints {
			cand := KnockoutPivotCands{
				PivotTime: pivotTime,
				User:      userAddr,
			}
			// Checking if cands already contains cand to not double knock out orders with multiple mints
			if (mint.PivotTime == pivotTime || isMintMaybeInPivot(mint.PivotTime, pivotTime, crossTime) || isMintMaybeInPivot(mint.TxTime, pivotTime, crossTime)) && !slices.Contains(cands, cand) {
				cands = append(cands, cand)
			}
		}
	}
	return cands
}

/* Returns true if there's a possibility that the minted liquidity may be knocked out on the
 * the pivot cross event. */
func isMintMaybeInPivot(mintTime int, pivotTime int, knockoutTime int) bool {
	return mintTime >= pivotTime && mintTime <= knockoutTime
}

func (k *KnockoutLiquiditySeries) GetActiveLiq() (activeLiq *big.Int) {
	k.lock.Lock()
	defer k.lock.Unlock()
	return big.NewInt(0).Set(&k.Active.ConcLiq)
}

func (k *KnockoutLiquiditySeries) UpdateActiveLiq(liqQty big.Int, refreshTime int64) {
	k.lock.Lock()
	defer k.lock.Unlock()
	k.Active.ConcLiq = liqQty
	k.Active.RefreshTime = refreshTime
}

func (k *KnockoutLiquiditySeries) UpdatePostKOLiq(pivotTime int, liqQty big.Int, refreshTime int64) {
	k.lock.Lock()
	defer k.lock.Unlock()
	posKoLiq, ok := k.KnockedOut[pivotTime]
	if !ok {
		posKoLiq = &PositionLiquidity{}
		k.KnockedOut[pivotTime] = posKoLiq
	}
	posKoLiq.ConcLiq = liqQty
	k.Active.RefreshTime = refreshTime
}
