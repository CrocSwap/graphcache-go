package cache

import (
	"bytes"

	"github.com/CrocSwap/graphcache-go/model"
	"github.com/CrocSwap/graphcache-go/types"
)

type MemoryCache struct {
	latestBlocks RWLockMap[types.ChainId, int64]

	userBalTokens RWLockMapArray[chainAndAddr, types.EthAddress]

	liqPosition   RWLockMap[types.PositionLocation, *model.PositionTracker]
	userPositions RWLockMapMap[chainAndAddr, types.PositionLocation, *model.PositionTracker]
	poolPositions RWLockMapMap[types.PoolLocation, types.PositionLocation, *model.PositionTracker]

	liqKnockouts  RWLockMap[types.PositionLocation, *model.KnockoutSubplot]
	userKnockouts RWLockMapMap[chainAndAddr, types.PositionLocation, *model.KnockoutSubplot]
	poolKnockouts RWLockMapMap[types.PoolLocation, types.PositionLocation, *model.KnockoutSubplot]

	knockoutSagas      RWLockMap[types.BookLocation, *model.KnockoutSaga]
	knockoutPivotTimes RWLockMap[types.BookLocation, int]

	userTxs        RWLockMapArray[chainAndAddr, types.PoolTxEvent]
	poolTxs        RWLockMapArray[types.PoolLocation, types.PoolTxEvent]
	poolPosUpdates RWLockMapArray[types.PoolLocation, PosAndLocPair]
	poolKoUpdates  RWLockMapArray[types.PoolLocation, KoAndLocPair]

	poolLiqCurve       RWLockMap[types.PoolLocation, *model.LiquidityCurve]
	poolTradingHistory RWLockMap[types.PoolLocation, *model.PoolTradingHistory]
	poolHourlyCandles  RWLockMap[types.PoolLocation, *[]model.Candle]
}

func New() *MemoryCache {
	return &MemoryCache{
		latestBlocks: newRwLockMap[types.ChainId, int64](),

		userBalTokens: newRwLockMapArray[chainAndAddr, types.EthAddress](),

		liqPosition:   newRwLockMap[types.PositionLocation, *model.PositionTracker](),
		userPositions: newRwLockMapMap[chainAndAddr, types.PositionLocation, *model.PositionTracker](),
		poolPositions: newRwLockMapMap[types.PoolLocation, types.PositionLocation, *model.PositionTracker](),

		liqKnockouts:  newRwLockMap[types.PositionLocation, *model.KnockoutSubplot](),
		userKnockouts: newRwLockMapMap[chainAndAddr, types.PositionLocation, *model.KnockoutSubplot](),
		poolKnockouts: newRwLockMapMap[types.PoolLocation, types.PositionLocation, *model.KnockoutSubplot](),

		knockoutSagas:      newRwLockMap[types.BookLocation, *model.KnockoutSaga](),
		knockoutPivotTimes: newRwLockMap[types.BookLocation, int](),

		userTxs:        newRwLockMapArray[chainAndAddr, types.PoolTxEvent](),
		poolTxs:        newRwLockMapArray[types.PoolLocation, types.PoolTxEvent](),
		poolPosUpdates: newRwLockMapArray[types.PoolLocation, PosAndLocPair](),
		poolKoUpdates:  newRwLockMapArray[types.PoolLocation, KoAndLocPair](),

		poolLiqCurve:       newRwLockMap[types.PoolLocation, *model.LiquidityCurve](),
		poolTradingHistory: newRwLockMap[types.PoolLocation, *model.PoolTradingHistory](),
		poolHourlyCandles:  newRwLockMap[types.PoolLocation, *[]model.Candle](),
	}
}

type chainAndAddr struct {
	types.ChainId
	types.EthAddress
}

type chainUserAndPool struct {
	user types.EthAddress
	types.PoolLocation
}

type PosAndLocPair struct {
	Loc types.PositionLocation
	Pos *model.PositionTracker
}

func (p PosAndLocPair) Time() int {
	return p.Pos.LatestUpdateTime
}

func (p PosAndLocPair) Hash(buf *bytes.Buffer) [32]byte {
	return p.Loc.Hash(buf)
}

type KoAndLocPair struct {
	Loc types.PositionLocation
	Ko  *model.KnockoutSubplot
}

func (k KoAndLocPair) Time() int {
	return k.Ko.LatestUpdateTime
}

func (k KoAndLocPair) Hash(buf *bytes.Buffer) [32]byte {
	return k.Loc.Hash(buf)
}
