package cache

import (
	"github.com/CrocSwap/graphcache-go/model"
	"github.com/CrocSwap/graphcache-go/types"
)

type MemoryCache struct {
	latestBlocks RWLockMap[types.ChainId, int64]

	userBalTokens RWLockMapArray[chainAndAddr, types.EthAddress]

	liqPosition          RWLockMap[types.PositionLocation, *model.PositionTracker]
	userPositions        RWLockMapMap[chainAndAddr, types.PositionLocation, *model.PositionTracker]
	poolPositions        RWLockMapMap[types.PoolLocation, types.PositionLocation, *model.PositionTracker]
	userAndPoolPositions RWLockMapMap[chainUserAndPool, types.PositionLocation, *model.PositionTracker]

	liqKnockouts         RWLockMap[types.PositionLocation, *model.KnockoutSubplot]
	userKnockouts        RWLockMapMap[chainAndAddr, types.PositionLocation, *model.KnockoutSubplot]
	poolKnockouts        RWLockMapMap[types.PoolLocation, types.PositionLocation, *model.KnockoutSubplot]
	userAndPoolKnockouts RWLockMapMap[chainUserAndPool, types.PositionLocation, *model.KnockoutSubplot]

	knockoutSagas RWLockMap[types.BookLocation, *model.KnockoutSaga]

	userTxs RWLockMapArray[chainAndAddr, types.PoolTxEvent]
	poolTxs RWLockMapArray[types.PoolLocation, types.PoolTxEvent]

	poolLiqCurve       RWLockMap[types.PoolLocation, *model.LiquidityCurve]
	poolTradingHistory RWLockMap[types.PoolLocation, *model.PoolTradingHistory]
}

func New() *MemoryCache {
	return &MemoryCache{
		latestBlocks: newRwLockMap[types.ChainId, int64](),

		userBalTokens: newRwLockMapArray[chainAndAddr, types.EthAddress](),

		liqPosition:          newRwLockMap[types.PositionLocation, *model.PositionTracker](),
		userPositions:        newRwLockMapMap[chainAndAddr, types.PositionLocation, *model.PositionTracker](),
		poolPositions:        newRwLockMapMap[types.PoolLocation, types.PositionLocation, *model.PositionTracker](),
		userAndPoolPositions: newRwLockMapMap[chainUserAndPool, types.PositionLocation, *model.PositionTracker](),

		liqKnockouts:         newRwLockMap[types.PositionLocation, *model.KnockoutSubplot](),
		userKnockouts:        newRwLockMapMap[chainAndAddr, types.PositionLocation, *model.KnockoutSubplot](),
		poolKnockouts:        newRwLockMapMap[types.PoolLocation, types.PositionLocation, *model.KnockoutSubplot](),
		userAndPoolKnockouts: newRwLockMapMap[chainUserAndPool, types.PositionLocation, *model.KnockoutSubplot](),

		knockoutSagas: newRwLockMap[types.BookLocation, *model.KnockoutSaga](),

		userTxs: newRwLockMapArray[chainAndAddr, types.PoolTxEvent](),
		poolTxs: newRwLockMapArray[types.PoolLocation, types.PoolTxEvent](),

		poolLiqCurve:       newRwLockMap[types.PoolLocation, *model.LiquidityCurve](),
		poolTradingHistory: newRwLockMap[types.PoolLocation, *model.PoolTradingHistory](),
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

type chainAndUserTokenAddr struct {
	types.ChainId
	user  types.EthAddress
	token types.EthAddress
}
