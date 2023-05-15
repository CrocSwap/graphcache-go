package models

import "github.com/CrocSwap/graphcache-go/types"

type MemoryCache struct {
	latestBlocks  RWLockMap[types.ChainId, int64]
	userBalTokens RWLockMapArray[chainAndAddr, types.EthAddress]
}

type chainAndAddr struct {
	chainId types.ChainId
	ethAddr types.EthAddress
}

func newMemCache() MemoryCache {
	return MemoryCache{
		latestBlocks:  newRwLockMap[types.ChainId, int64](),
		userBalTokens: newRwLockMapArray[chainAndAddr, types.EthAddress](),
	}
}
