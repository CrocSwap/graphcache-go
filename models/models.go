package models

import "github.com/CrocSwap/graphcache-go/types"

type Models struct {
	cache MemoryCache
}

func New() *Models {
	return &Models{
		cache: newMemCache(),
	}
}

func (m *Models) LatestBlock(chainId types.ChainId) int64 {
	block, okay := m.cache.latestBlocks.lookup(chainId)
	if okay {
		return block
	} else {
		return -1
	}
}

func (m *Models) RetrieveUserBalances(chainId types.ChainId, user: types.EthAddress) []types.UserBalance {
	key := chainAndAddr{chainId: chainId, ethAddr: user}
	bals, okay := m.cache.userBals.lookup(key)
	if okay {
		return bals
	} else {
		return make([]types.UserBalance, )
	}
}
