package views

import (
	"log"
	"slices"
	"time"

	"github.com/CrocSwap/graphcache-go/model"
	"github.com/CrocSwap/graphcache-go/types"
)

type PoolStats struct {
	*AdditionalPoolStatsFields // Optional, only set when all pool stats are requested
	model.AccumPoolStats
	InitTime int `json:"initTime"`
	Events   int `json:"events"`
}

type AdditionalPoolStatsFields struct {
	*types.PoolLocation
	PriceSwap24hAgo   *float64 `json:"priceSwap24hAgo,omitempty"`
	PriceLiq24hAgo    *float64 `json:"priceLiq24hAgo,omitempty"`
	PriceIndic24hAgo  *float64 `json:"priceIndic24hAgo,omitempty"`
	BaseVolume24hAgo  *float64 `json:"baseVolume24hAgo,omitempty"`
	QuoteVolume24hAgo *float64 `json:"quoteVolume24hAgo,omitempty"`
	BaseFees24hAgo    *float64 `json:"baseFees24hAgo,omitempty"`
	QuoteFees24hAgo   *float64 `json:"quoteFees24hAgo,omitempty"`
}

func (v *Views) QueryPoolStats(chainId types.ChainId,
	base types.EthAddress, quote types.EthAddress, poolIdx int, with24hPrices bool) PoolStats {

	loc := types.PoolLocation{
		ChainId: chainId,
		PoolIdx: poolIdx,
		Base:    base,
		Quote:   quote,
	}

	accum, eventCount := v.Cache.RetrievePoolAccum(loc)
	firstAccum := v.Cache.RetrievePoolAccumFirst(loc)

	stats := PoolStats{
		InitTime:       firstAccum.LatestTime,
		AccumPoolStats: accum,
		Events:         eventCount,
	}

	if with24hPrices {
		stats24hAgo := v.QueryPoolStatsFrom(chainId, base, quote, poolIdx, int(time.Now().Unix())-24*3600)
		stats.AdditionalPoolStatsFields = &AdditionalPoolStatsFields{
			PriceSwap24hAgo:   &stats24hAgo.LastPriceSwap,
			PriceLiq24hAgo:    &stats24hAgo.LastPriceLiq,
			PriceIndic24hAgo:  &stats24hAgo.LastPriceIndic,
			BaseVolume24hAgo:  &stats24hAgo.BaseVolume,
			QuoteVolume24hAgo: &stats24hAgo.QuoteVolume,
			BaseFees24hAgo:    &stats24hAgo.BaseFees,
			QuoteFees24hAgo:   &stats24hAgo.QuoteFees,
		}
	}

	return stats
}

func (v *Views) QueryPoolStatsFrom(chainId types.ChainId,
	base types.EthAddress, quote types.EthAddress, poolIdx int, histTime int) PoolStats {

	loc := types.PoolLocation{
		ChainId: chainId,
		PoolIdx: poolIdx,
		Base:    base,
		Quote:   quote,
	}
	accum, eventCount := v.Cache.RetrievePoolAccumBefore(loc, histTime)
	firstAccum := v.Cache.RetrievePoolAccumFirst(loc)

	return PoolStats{
		InitTime:       firstAccum.LatestTime,
		AccumPoolStats: accum,
		Events:         eventCount,
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

	startTime = startTime - startTime%timeRange.Period
	endTime = endTime - endTime%timeRange.Period
	// log.Println("QueryPoolCandles", timeRange, startTime, endTime)

	if timeRange.Period >= 3600 && timeRange.Period%3600 == 0 {
		return v.fastHourlyCandles(loc, timeRange, startTime, endTime)
	}

	start := time.Now()
	open, series := v.Cache.RetrievePoolAccumSeries(loc, startTime, endTime)
	diff := time.Since(start)
	if diff > 250*time.Millisecond {
		log.Println("Slow RetrievePoolAccumSeries:", diff)
	}

	// log.Println("open", open, "series", len(series), "from", startTime, "to", endTime)

	start = time.Now()
	builder := model.NewCandleBuilder(startTime, timeRange.Period, open)
	for _, accum := range series {
		builder.Increment(accum)
	}
	diff = time.Since(start)
	if diff > 25*time.Millisecond {
		log.Println("Slow buildCandles:", diff)
	}
	candles := builder.Close(endTime)
	return candles
}

// Uses a per pool cache of hourly candles to build 1h/4h/1d/whatever candles.
func (v *Views) fastHourlyCandles(loc types.PoolLocation, timeRange CandleRangeArgs, startTime int, endTime int) []model.Candle {
	candlesPtr, candleLock := v.Cache.BorrowPoolHourlyCandles(loc, false)
	candles := *candlesPtr
	if len(candles) > 0 {
		// log.Println("Cached candles", len(candles), candles[0].Time, candles[len(candles)-1].Time)
	}
	// If no cached candles or the last candle is stale, refresh
	if len(candles) == 0 || (len(candles) > 0 && time.Since(time.Unix(int64(candles[len(candles)-1].Time), 0)) > 2*time.Hour+1*time.Minute) {
		pos, poolLock := v.Cache.BorrowPoolTradingHist(loc, false)
		if pos == nil || len(pos.TimeSnaps) == 0 {
			if pos != nil {
				poolLock.RUnlock()
			}
			return []model.Candle{}
		}
		firstHourly := pos.TimeSnaps[0].LatestTime - pos.TimeSnaps[0].LatestTime%3600
		now := int(time.Now().Unix())
		lastHourly := now - now%3600 + 1
		log.Println("Building hourly candles for", loc.Base, loc.Quote, firstHourly, lastHourly, "from", pos.TimeSnaps[0].LatestTime, pos.TimeSnaps[len(pos.TimeSnaps)-1].LatestTime)
		builder := model.NewCandleBuilder(firstHourly, 3600, pos.TimeSnaps[0])
		for _, accum := range pos.TimeSnaps {
			if accum.LatestTime < lastHourly {
				builder.Increment(accum)
			}
		}
		if pos.StatsCounter.LatestTime < lastHourly {
			builder.Increment(pos.StatsCounter)
		}
		poolLock.RUnlock()
		candles = builder.Close(lastHourly)
		if len(candles) > 0 {
			log.Println("Built candles", len(candles), candles[0].Time, candles[len(candles)-1].Time)
		}
		*candlesPtr = candles
	}
	if len(candles) > 0 {
		// log.Println("CombineHourlyCandles", candles[0].Time, candles[len(candles)-1].Time, startTime, endTime)
	}
	candles = model.CombineHourlyCandles(candles, timeRange.Period/3600, startTime, endTime, timeRange.N)
	candleLock.RUnlock()

	// Since the last candle may include the latest data, it needs to be generated
	if endTime > int(time.Now().Unix()) {
		start := time.Now()
		open, series := v.Cache.RetrievePoolAccumSeries(loc, endTime-timeRange.Period*2, endTime-timeRange.Period)
		diff := time.Since(start)
		if diff > 250*time.Millisecond {
			log.Println("Slow fastRetrievePoolAccumSeries:", diff)
		}
		// log.Println("fastRetrievePoolAccumSeries", len(series), open, endTime-timeRange.Period*2, endTime-timeRange.Period)
		if len(series) > 0 {
			// log.Println(series[0].LatestTime, "---", series[len(series)-1].LatestTime)
		}

		start = time.Now()
		builder := model.NewCandleBuilder(endTime-timeRange.Period*2, timeRange.Period, open)
		for _, accum := range series {
			builder.Increment(accum)
		}
		// if pos.StatsCounter.LatestTime < lastHourly {
		// 	builder.Increment(pos.StatsCounter)
		// }
		diff = time.Since(start)
		if diff > 25*time.Millisecond {
			log.Println("Slow fastBuildCandles:", diff)
		}
		lastCandle := builder.Close(endTime)
		// log.Println("Last candle fast", candles[len(candles)-1])
		// log.Println("Last candle new ", lastCandle)
		if len(lastCandle) > 0 {
			if len(candles) > 0 && lastCandle[0].Time == candles[len(candles)-1].Time {
				candles[len(candles)-1] = lastCandle[0]
			} else {
				candles = append(candles, lastCandle[0])
			}
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

func (v *Views) QueryAllPoolStats(chainId types.ChainId, histTime int, with24hPrices bool) (result []PoolStats) {
	fullSet := v.Cache.RetrievePoolSet()

	for _, pool := range fullSet {
		if pool.ChainId == chainId {
			var stats PoolStats
			if histTime <= 0 {
				stats = v.QueryPoolStats(pool.ChainId, pool.Base, pool.Quote, pool.PoolIdx, with24hPrices)
			} else {
				stats = v.QueryPoolStatsFrom(pool.ChainId, pool.Base, pool.Quote, pool.PoolIdx, histTime)
			}
			if stats.AdditionalPoolStatsFields == nil {
				stats.AdditionalPoolStatsFields = &AdditionalPoolStatsFields{}
			}
			stats.PoolLocation = &pool
			result = append(result, stats)
		}
	}

	slices.SortFunc(result, func(i, j PoolStats) int {
		if i.Events != j.Events {
			return j.Events - i.Events
		}
		if i.InitTime != j.InitTime {
			return i.InitTime - j.InitTime
		}
		if i.Base < j.Base {
			return -1
		} else if i.Base > j.Base {
			return 1
		} else {
			if i.Quote < j.Quote {
				return -1
			} else if i.Quote > j.Quote {
				return 1
			} else {
				return i.PoolIdx - j.PoolIdx
			}
		}
	})

	return
}
