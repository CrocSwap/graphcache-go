package models

import (
	"github.com/CrocSwap/graphcache-go/types"
)

type Models struct {
	cache    MemoryCache
	watchers mutateWatchers
}

func New() *Models {
	return &Models{
		cache:    newMemCache(),
		watchers: MutateWatchers(),
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
	key := chainAndAddr{chainId, user}
	tokens, okay := m.cache.userBalTokens.lookup(key)
	if okay {
		return tokens
	} else {
		return make([]types.EthAddress, 0)
	}
}

func (m *Models) RetrieveUserPositions(chainId types.ChainId, user types.EthAddress) map[types.PositionLocation]*PositionTracker {
	key := chainAndAddr{chainId, user}
	pos, okay := m.cache.userPositions.lookupSet(key)
	if okay {
		return pos
	} else {
		return make(map[types.PositionLocation]*PositionTracker)
	}
}

func (m *Models) AddUserBalance(chainId types.ChainId, user types.EthAddress, token types.EthAddress) {
	key := chainAndAddr{chainId, user}
	m.cache.userBalTokens.insert(key, token)
}

func (m *Models) UpdatePositionMint(loc types.PositionLocation, time int) {
	val := m.materializePosition(loc)
	m.watchers.positions <- posUpdateMsg{pos: val, update: posMint, time: time}
}

func (m *Models) UpdatePositionBurn(loc types.PositionLocation, time int) {
	val := m.materializePosition(loc)
	m.watchers.positions <- posUpdateMsg{pos: val, update: posBurn, time: time}
}

func (m *Models) UpdatePositionHarvest(loc types.PositionLocation, time int) {
	val := m.materializePosition(loc)
	m.watchers.positions <- posUpdateMsg{pos: val, update: posHarvest, time: time}
}

func (m *Models) materializePosition(loc types.PositionLocation) *PositionTracker {
	val, ok := m.cache.liqPosition.lookup(loc)
	if !ok {
		val = &PositionTracker{}
		m.cache.liqPosition.insert(loc, val)
		m.cache.userPositions.insert(chainAndAddr{loc.ChainId, loc.User}, loc, val)
		m.cache.poolPositions.insert(chainAndPool{loc.ChainId, loc.PoolLocation}, loc, val)
		m.cache.userAndPoolPositions.insert(
			chainUserAndPool{loc.ChainId, loc.User, loc.PoolLocation}, loc, val)
	}
	return val
}
