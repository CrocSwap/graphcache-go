package models

import "github.com/CrocSwap/graphcache-go/types"

type MemoryCache struct {
	latestBlocks  RWLockMap[types.ChainId, int64]
	userBalTokens RWLockMapArray[chainAndAddr, types.EthAddress]

	liqPosition          RWLockMap[types.PositionLocation, *PositionTracker]
	userPositions        RWLockMapMap[chainAndAddr, types.PositionLocation, *PositionTracker]
	poolPositions        RWLockMapMap[chainAndPool, types.PositionLocation, *PositionTracker]
	userAndPoolPositions RWLockMapMap[chainUserAndPool, types.PositionLocation, *PositionTracker]
}

func newMemCache() MemoryCache {
	return MemoryCache{
		latestBlocks:         newRwLockMap[types.ChainId, int64](),
		userBalTokens:        newRwLockMapArray[chainAndAddr, types.EthAddress](),
		liqPosition:          newRwLockMap[types.PositionLocation, *PositionTracker](),
		userPositions:        newRwLockMapMap[chainAndAddr, types.PositionLocation, *PositionTracker](),
		poolPositions:        newRwLockMapMap[chainAndPool, types.PositionLocation, *PositionTracker](),
		userAndPoolPositions: newRwLockMapMap[chainUserAndPool, types.PositionLocation, *PositionTracker](),
	}
}

type chainAndAddr struct {
	types.ChainId
	types.EthAddress
}

type chainAndPool struct {
	types.ChainId
	types.PoolLocation
}

type chainUserAndPool struct {
	types.ChainId
	types.EthAddress
	types.PoolLocation
}

type chainAndUserTokenAddr struct {
	types.ChainId
	user  types.EthAddress
	token types.EthAddress
}
