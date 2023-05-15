package models

import "github.com/CrocSwap/graphcache-go/types"

type MemoryCache struct {
	latestBlocks  RWLockMap[types.ChainId, int64]
	userBalTokens RWLockMap[chainAndAddr, []types.EthAddress]
}

type chainAndAddr struct {
	chainId types.ChainId
	ethAddr types.EthAddress
}

func newMemCache() MemoryCache {
	return MemoryCache{
		latestBlocks:  newRwLockMap[types.ChainId, int64](),
		userBalTokens: newRwLockMap[chainAndAddr, []types.EthAddress](),
	}
}
