package views

import (
	"github.com/CrocSwap/graphcache-go/cache"
	"github.com/CrocSwap/graphcache-go/loader"
	"github.com/CrocSwap/graphcache-go/model"
	"github.com/CrocSwap/graphcache-go/types"
)

type IViews interface {
	QueryUserTokens(chainId types.ChainId, user types.EthAddress) UserTokensResponse

	QueryUserPositions(chainId types.ChainId, user types.EthAddress) []UserPosition
	QueryPoolPositions(chainId types.ChainId, base types.EthAddress, quote types.EthAddress,
		poolIdx int, nResults int, omitEmpty bool, afterTime int, beforeTime int) []UserPosition
	QueryPoolApyLeaders(chainId types.ChainId, base types.EthAddress, quote types.EthAddress,
		poolIdx int, nResults int, omitEmpty bool) []UserPosition
	QueryUserPoolPositions(chainId types.ChainId, user types.EthAddress,
		base types.EthAddress, quote types.EthAddress, poolIdx int) []UserPosition
	QuerySinglePosition(chainId types.ChainId, user types.EthAddress,
		base types.EthAddress, quote types.EthAddress,
		poolIdx int, bidTick int, askTick int) *UserPosition
	QueryHistoricPositions(chainId types.ChainId, base types.EthAddress, quote types.EthAddress,
		poolIdx int, time int, user types.EthAddress, omitEmpty bool) []HistoricUserPosition

	QueryUserLimits(chainId types.ChainId, user types.EthAddress) []UserLimitOrder
	QueryPoolLimits(chainId types.ChainId, base types.EthAddress, quote types.EthAddress, poolIdx int,
		nResults int, afterTime int, beforeTime int) []UserLimitOrder
	QueryUserPoolLimits(chainId types.ChainId, user types.EthAddress,
		base types.EthAddress, quote types.EthAddress, poolIdx int) []UserLimitOrder
	QueryUserPoolTxHist(chainId types.ChainId, user types.EthAddress,
		base types.EthAddress, quote types.EthAddress, poolIdx int,
		nResults int, afterTime int, beforeTime int) []UserTxHistory
	QuerySingleLimit(chainId types.ChainId, user types.EthAddress,
		base types.EthAddress, quote types.EthAddress,
		poolIdx int, bidTick int, askTick int, isBid bool, pivotTime int) *UserLimitOrder

	QueryUserTxHist(chainId types.ChainId, user types.EthAddress,
		nResults int, afterTime int, beforeTime int) []UserTxHistory
	QueryPoolTxHist(chainId types.ChainId, base types.EthAddress, quote types.EthAddress, poolIdx int,
		nResults int, afterTime int, beforeTime int) []UserTxHistory
	QueryPoolLiquidityCurve(chainId types.ChainId, base types.EthAddress, quote types.EthAddress,
		poolIdx int) PoolLiqCurve

	QueryPoolStats(chainId types.ChainId, base types.EthAddress, quote types.EthAddress,
		poolIdx int, with24hPrices bool) PoolStats
	QueryPoolStatsFrom(chainId types.ChainId, base types.EthAddress, quote types.EthAddress,
		poolIdx int, histTime int) PoolStats
	QueryAllPoolStats(chainId types.ChainId, histTime int, with24hPrices bool) []PoolStats
	QueryChainStats(chainId types.ChainId, nResults int) []TokenDexAgg

	QueryPoolCandles(chainId types.ChainId, base types.EthAddress, quote types.EthAddress, poolIdx int,
		timeRange CandleRangeArgs) []model.Candle

	QueryPoolSet(chainId types.ChainId) []types.PoolLocation

	QueryPlumeUserTask(user types.EthAddress, task string) PlumeTaskStatus
}

type Views struct {
	Cache   *cache.MemoryCache
	OnChain *loader.OnChainLoader
}
