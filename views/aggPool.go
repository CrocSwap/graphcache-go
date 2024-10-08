package views

import (
	"time"

	"github.com/CrocSwap/graphcache-go/model"
	"github.com/CrocSwap/graphcache-go/types"
)

type PoolStats struct {
	InitTime int `json:"initTime"`
	model.AccumPoolStats
}

func (v *Views) QueryPoolStats(chainId types.ChainId,
	base types.EthAddress, quote types.EthAddress, poolIdx int) PoolStats {

	loc := types.PoolLocation{
		ChainId: chainId,
		PoolIdx: poolIdx,
		Base:    base,
		Quote:   quote,
	}

	accum := v.Cache.RetrievePoolAccum(loc)
	firstAccum := v.Cache.RetrievePoolAccumFirst(loc)

	return PoolStats{
		InitTime:       firstAccum.LatestTime,
		AccumPoolStats: accum,
	}
}

func (v *Views) QueryPoolStatsFrom(chainId types.ChainId,
	base types.EthAddress, quote types.EthAddress, poolIdx int, histTime int) PoolStats {

	loc := types.PoolLocation{
		ChainId: chainId,
		PoolIdx: poolIdx,
		Base:    base,
		Quote:   quote,
	}
	accum := v.Cache.RetrievePoolAccumBefore(loc, histTime)
	firstAccum := v.Cache.RetrievePoolAccumFirst(loc)

	return PoolStats{
		InitTime:       firstAccum.LatestTime,
		AccumPoolStats: accum,
	}
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
		startTime = *timeRange.StartTime
		endTime = startTime + timeRange.N*timeRange.Period
	}

	open, series := v.Cache.RetrievePoolAccumSeries(loc, startTime, endTime)

	builder := model.NewCandleBuilder(startTime, timeRange.Period, open)
	for _, accum := range series {
		builder.Increment(accum)
	}
	return builder.Close(endTime)
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
