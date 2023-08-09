package views

import (
	"time"

	"github.com/CrocSwap/graphcache-go/model"
	"github.com/CrocSwap/graphcache-go/types"
	"github.com/CrocSwap/graphcache-go/utils"
)

var uniswapCandles = utils.GoDotEnvVariable("UNISWAP_CANDLES") == "true"
var stdDevWindowSize = utils.GetEnvVarIntFromString("UNISWAP_STD_DEV_WINDOW_SIZE", 5)

func (v *Views) QueryPoolStats(chainId types.ChainId,
	base types.EthAddress, quote types.EthAddress, poolIdx int) model.AccumPoolStats {

	loc := types.PoolLocation{
		ChainId: chainId,
		PoolIdx: poolIdx,
		Base:    base,
		Quote:   quote,
	}
	return v.Cache.RetrievePoolAccum(loc)
}

func (v *Views) QueryPoolStatsFrom(chainId types.ChainId,
	base types.EthAddress, quote types.EthAddress, poolIdx int, histTime int) model.AccumPoolStats {

	loc := types.PoolLocation{
		ChainId: chainId,
		PoolIdx: poolIdx,
		Base:    base,
		Quote:   quote,
	}
	return v.Cache.RetrievePoolAccumBefore(loc, histTime)
}

type CandleRangeArgs struct {
	N         int  // Number of candles
	Period    int  // Candle size in seconds
	StartTime *int // If nil serve most recent
}

func (v *Views) QueryPoolCandles(chainId types.ChainId, base types.EthAddress, quote types.EthAddress, poolIdx int,
	timeRange CandleRangeArgs) []model.Candle {
	loc := types.PoolLocation{
		ChainId: chainId,
		PoolIdx: poolIdx,
		Base:    base,
		Quote:   quote,
	}

	endTime := int(time.Now().Unix())
	startTime := endTime - timeRange.N*timeRange.Period

	if timeRange.StartTime != nil {

		if(uniswapCandles){
            endTime = *timeRange.StartTime + timeRange.Period
            startTime = *timeRange.StartTime - timeRange.N *timeRange.Period
        }else {
            startTime = *timeRange.StartTime
            endTime = startTime + timeRange.N*timeRange.Period
        }
	}
	open, series := v.Cache.RetrievePoolAccumSeries(loc, startTime, endTime)

	rollingStdDev  := model.ComputeRollingStdDev(series, stdDevWindowSize)

	builder := model.NewCandleBuilder(startTime, timeRange.Period, open)
	for _, accum := range series {
		builder.Increment(accum, rollingStdDev)
	}
	 candles := builder.Close(endTime)
	 

	 if(uniswapCandles){
		//  Reverse the order of the candles
		for i, j := 0, len(candles)-1; i < j; i, j = i+1, j-1 {
			candles[i], candles[j] = candles[j], candles[i]
		}
	}
	return candles
}

func (v *Views) QueryPoolSet(chainId types.ChainId) []types.PoolLocation {
	fullSet := v.Cache.RetrievePoolSet()

	poolSet := make([]types.PoolLocation, 0)
	for _, pool := range fullSet {
		if pool.ChainId == chainId {
			poolSet = append(poolSet, pool)
		}
	}

	return poolSet
}
