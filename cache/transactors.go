package cache

import (
	"github.com/CrocSwap/graphcache-go/model"
	"github.com/CrocSwap/graphcache-go/types"
)

func (m *MemoryCache) LatestBlock(chainId types.ChainId) int64 {
	block, okay := m.latestBlocks.lookup(chainId)
	if okay {
		return block
	} else {
		return -1
	}
}

func (m *MemoryCache) RetrieveUserBalances(chainId types.ChainId, user types.EthAddress) []types.EthAddress {
	key := chainAndAddr{chainId, user}
	tokens, okay := m.userBalTokens.lookup(key)
	if okay {
		return tokens
	} else {
		return make([]types.EthAddress, 0)
	}
}

func (m *MemoryCache) RetrieveUserPositions(
	chainId types.ChainId,
	user types.EthAddress) map[types.PositionLocation]*model.PositionTracker {
	key := chainAndAddr{chainId, user}
	pos, okay := m.userPositions.lookupSet(key)
	if okay {
		return pos
	} else {
		return make(map[types.PositionLocation]*model.PositionTracker)
	}
}

func (m *MemoryCache) AddUserBalance(chainId types.ChainId, user types.EthAddress, token types.EthAddress) {
	key := chainAndAddr{chainId, user}
	m.userBalTokens.insert(key, token)
}

func (m *MemoryCache) MaterializePosition(loc types.PositionLocation) *model.PositionTracker {
	val, ok := m.liqPosition.lookup(loc)
	if !ok {
		val = &model.PositionTracker{}
		m.liqPosition.insert(loc, val)
		m.userPositions.insert(chainAndAddr{loc.ChainId, loc.User}, loc, val)
		m.poolPositions.insert(chainAndPool{loc.ChainId, loc.PoolLocation}, loc, val)
		m.userAndPoolPositions.insert(
			chainUserAndPool{loc.ChainId, loc.User, loc.PoolLocation}, loc, val)
	}
	return val
}
