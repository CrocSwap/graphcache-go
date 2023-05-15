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

func (m *Models) RetrieveUserBalances(chainId types.ChainId, user types.EthAddress) []types.EthAddress {
	key := chainAndAddr{chainId: chainId, ethAddr: user}
	tokens, okay := m.cache.userBalTokens.lookup(key)
	if okay {
		return tokens
	} else {
		return make([]types.EthAddress, 0)
	}
}
