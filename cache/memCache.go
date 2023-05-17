package cache

import (
	"github.com/CrocSwap/graphcache-go/model"
	"github.com/CrocSwap/graphcache-go/types"
)

type MemoryCache struct {
	latestBlocks  RWLockMap[types.ChainId, int64]
	tokenMetadata RWLockMap[chainAndAddr, *model.TokenMetadataHandle]

	userBalTokens RWLockMapArray[chainAndAddr, types.EthAddress]

	liqPosition          RWLockMap[types.PositionLocation, *model.PositionTracker]
	userPositions        RWLockMapMap[chainAndAddr, types.PositionLocation, *model.PositionTracker]
	poolPositions        RWLockMapMap[chainAndPool, types.PositionLocation, *model.PositionTracker]
	userAndPoolPositions RWLockMapMap[chainUserAndPool, types.PositionLocation, *model.PositionTracker]
}

func New() *MemoryCache {
	return &MemoryCache{
		latestBlocks:         newRwLockMap[types.ChainId, int64](),
		tokenMetadata:        newRwLockMap[chainAndAddr, *model.TokenMetadataHandle](),
		userBalTokens:        newRwLockMapArray[chainAndAddr, types.EthAddress](),
		liqPosition:          newRwLockMap[types.PositionLocation, *model.PositionTracker](),
		userPositions:        newRwLockMapMap[chainAndAddr, types.PositionLocation, *model.PositionTracker](),
		poolPositions:        newRwLockMapMap[chainAndPool, types.PositionLocation, *model.PositionTracker](),
		userAndPoolPositions: newRwLockMapMap[chainUserAndPool, types.PositionLocation, *model.PositionTracker](),
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
