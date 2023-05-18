package model

import (
	"github.com/CrocSwap/graphcache-go/loader"
	"github.com/CrocSwap/graphcache-go/types"
)

func InitPoolState(onChain *loader.OnChainLoader, loc types.PoolLocation) *ExpiryHandle[types.PoolPriceLiq] {
	fetchFn := func() types.PoolPriceLiq {
		pool := types.PoolPriceLiq{}
		return pool
	}

	return InitCacheHandle(fetchFn, POOL_PRICE_TIMEOUT, types.PoolPriceLiq{})
}

const POOL_PRICE_TIMEOUT = 30
