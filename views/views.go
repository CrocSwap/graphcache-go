package views

import (
	"github.com/CrocSwap/graphcache-go/cache"
	"github.com/CrocSwap/graphcache-go/loader"
	"github.com/CrocSwap/graphcache-go/types"
)

type IViews interface {
	QueryUserTokens(chainId types.ChainId, user types.EthAddress) UserTokensResponse

	QueryUserPositions(chainId types.ChainId, user types.EthAddress) []UserPosition
	QueryPoolPositions(chainId types.ChainId, base types.EthAddress, quote types.EthAddress,
		poolIdx int, nResults int, omitEmpty bool) []UserPosition
	QueryPoolApyLeaders(chainId types.ChainId, base types.EthAddress, quote types.EthAddress,
		poolIdx int, nResults int, omitEmpty bool) []UserPosition
	QueryUserPoolPositions(chainId types.ChainId, user types.EthAddress,
		base types.EthAddress, quote types.EthAddress,
		poolIdx int) []UserPosition
	QuerySinglePosition(chainId types.ChainId, user types.EthAddress,
		base types.EthAddress, quote types.EthAddress,
		poolIdx int, bidTick int, askTick int) *UserPosition

	QueryUserLimits(chainId types.ChainId, user types.EthAddress) []UserLimitOrder
	QueryPoolLimits(chainId types.ChainId, base types.EthAddress, quote types.EthAddress,
		poolIdx int, nResults int) []UserLimitOrder
	QueryUserPoolLimits(chainId types.ChainId, user types.EthAddress,
		base types.EthAddress, quote types.EthAddress,
		poolIdx int) []UserLimitOrder
	QuerySingleLimit(chainId types.ChainId, user types.EthAddress,
		base types.EthAddress, quote types.EthAddress,
		poolIdx int, bidTick int, askTick int, isBid bool, pivotTime int) *UserLimitOrder

	QueryUserTxHist(chainId types.ChainId, user types.EthAddress, nResults int) []UserTxHistory
	QueryPoolTxHist(chainId types.ChainId, base types.EthAddress, quote types.EthAddress,
		poolIdx int, nResults int) []UserTxHistory
	QueryPoolTxHistFrom(chainId types.ChainId, base types.EthAddress, quote types.EthAddress,
		poolIdx int, nResults int, time int, period int) []UserTxHistory
}

type Views struct {
	Cache   *cache.MemoryCache
	OnChain *loader.OnChainLoader
}

const MAX_POOL_POSITIONS = 100
