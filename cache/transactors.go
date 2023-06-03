package cache

import (
	"github.com/CrocSwap/graphcache-go/loader"
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

func (m *MemoryCache) MaterializeTokenMetata(onChain *loader.OnChainLoader,
	chainId types.ChainId, token types.EthAddress) *model.ExpiryHandle[types.TokenMetadata] {

	key := chainAndAddr{chainId, token}
	hndl, okay := m.tokenMetadata.lookup(key)
	if !okay {
		hndl = model.InitTokenMetadata(onChain, chainId, token)
		m.tokenMetadata.insert(key, hndl)
	}
	return hndl
}

func (m *MemoryCache) MaterializePoolPrice(onChain *loader.OnChainLoader,
	loc types.PoolLocation) *model.ExpiryHandle[types.PoolPriceLiq] {

	hndl, okay := m.poolPrices.lookup(loc)
	if !okay {
		hndl = model.InitPoolState(onChain, loc)
		m.poolPrices.insert(loc, hndl)
	}
	return hndl
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

func (m *MemoryCache) RetrievePoolPositions(loc types.PoolLocation) map[types.PositionLocation]*model.PositionTracker {
	pos, okay := m.poolPositions.lookupSet(loc)
	if okay {
		return pos
	} else {
		return make(map[types.PositionLocation]*model.PositionTracker)
	}
}

func (m *MemoryCache) RetrieveUserPoolPositions(user types.EthAddress, pool types.PoolLocation) map[types.PositionLocation]*model.PositionTracker {
	loc := chainUserAndPool{user, pool}
	pos, okay := m.userAndPoolPositions.lookupSet(loc)
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
		m.poolPositions.insert(loc.PoolLocation, loc, val)
		m.userAndPoolPositions.insert(
			chainUserAndPool{loc.User, loc.PoolLocation}, loc, val)
	}
	return val
}

func (m *MemoryCache) MaterializeKnockoutBook(loc types.BookLocation) *model.KnockoutSaga {
	val, ok := m.knockoutSagas.lookup(loc)
	if !ok {
		val = model.NewKnockoutSaga()
		m.knockoutSagas.insert(loc, val)
	}
	return val
}

func (m *MemoryCache) MaterializeKnockoutPos(loc types.PositionLocation) *model.KnockoutSubplot {
	val, ok := m.liqKnockouts.lookup(loc)
	if !ok {
		saga := m.MaterializeKnockoutBook(loc.ToBookLoc())
		val = saga.ForUser(loc.User)
		m.liqKnockouts.insert(loc, val)
		m.userKnockouts.insert(chainAndAddr{loc.ChainId, loc.User}, loc, val)
		m.poolKnockouts.insert(loc.PoolLocation, loc, val)
		m.userAndPoolKnockouts.insert(
			chainUserAndPool{loc.User, loc.PoolLocation}, loc, val)
	}
	return val
}
