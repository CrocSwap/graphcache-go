package views

import (
	"github.com/CrocSwap/graphcache-go/cache"
	"github.com/CrocSwap/graphcache-go/loader"
	"github.com/CrocSwap/graphcache-go/types"
)

type IViews interface {
	QueryUserTokens(chainId types.ChainId, user types.EthAddress) (UserTokensResponse, error)

	QueryUserPositions(chainId types.ChainId, user types.EthAddress) ([]UserPosition, error)
	QueryPoolPositions(chainId types.ChainId, base types.EthAddress, quote types.EthAddress,
		poolIdx int, nResults int, omitEmpty bool) ([]UserPosition, error)
	QueryUserPoolPositions(chainId types.ChainId, user types.EthAddress,
		base types.EthAddress, quote types.EthAddress,
		poolIdx int) ([]UserPosition, error)
	QuerySinglePosition(chainId types.ChainId, user types.EthAddress,
		base types.EthAddress, quote types.EthAddress,
		poolIdx int, bidTick int, askTick int) (*UserPosition, error)

	QueryUserLimits(chainId types.ChainId, user types.EthAddress) ([]UserLimitOrder, error)
	QueryPoolLimits(chainId types.ChainId, base types.EthAddress, quote types.EthAddress,
		poolIdx int, nResults int) ([]UserLimitOrder, error)
	QueryUserPoolLimits(chainId types.ChainId, user types.EthAddress,
		base types.EthAddress, quote types.EthAddress,
		poolIdx int) ([]UserLimitOrder, error)
	QuerySingleLimit(chainId types.ChainId, user types.EthAddress,
		base types.EthAddress, quote types.EthAddress,
		poolIdx int, bidTick int, askTick int, isBid bool, pivotTime int) (*UserLimitOrder, error)
}

type Views struct {
	Cache   *cache.MemoryCache
	OnChain *loader.OnChainLoader
}
