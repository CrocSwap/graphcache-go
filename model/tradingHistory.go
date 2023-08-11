package model

import (
	"fmt"
	"math"

	"github.com/CrocSwap/graphcache-go/tables"
	"github.com/montanaflynn/stats"
)

type PoolTradingHistory struct {
	StatsCounter AccumPoolStats
	TimeSnaps    []AccumPoolStats
}

func NewPoolTradingHistory() *PoolTradingHistory {
	return &PoolTradingHistory{
		StatsCounter: AccumPoolStats{
			BaseTvl:        0.0,
			QuoteTvl:       0.0,
			BaseVolume:     0.0,
			QuoteVolume:    0.0,
			BaseFees:       0.0,
			QuoteFees:      0.0,
			LastPriceSwap:  0.0,
			LastPriceIndic: 0.0,
			LastPriceLiq:   0.0,
			FeeRate:        0.0,
		},
		TimeSnaps: make([]AccumPoolStats, 0),
	}
}

func (h *PoolTradingHistory) NextEvent(r tables.AggEvent) {
	if r.Time != h.StatsCounter.LatestTime {
		h.TimeSnaps = append(h.TimeSnaps, h.StatsCounter)
	}
	h.StatsCounter.Accumulate(r)
}

type RollingStdDev map[int]float64 

func ComputeRollingMAD(data []AccumPoolStats, windowSize int) RollingStdDev {
	if(len(data) < windowSize){
		return RollingStdDev{}
	}

	prices := make([]float64, len(data))

	for i, d := range data {
		prices[i] = float64(d.LastPriceSwap)
	}
	rangeStdDev := make([]float64, len(data)-windowSize+1)

	rollingStdDev := make(RollingStdDev)

	for i := range rangeStdDev {
		sdev, err := stats.MedianAbsoluteDeviation(prices[i:i+windowSize])
		if err != nil {
			fmt.Println("Rolling Standard Dev error")
		}
		rollingStdDev[data[i].LatestTime] = sdev
	}


	return rollingStdDev
}

type AccumPoolStats struct {
	LatestTime     int     `json:"latestTime"`
	BaseTvl        float64 `json:"baseTvl"`
	QuoteTvl       float64 `json:"quoteTvl"`
	BaseVolume     float64 `json:"baseVolume"`
	QuoteVolume    float64 `json:"quoteVolume"`
	BaseFees       float64 `json:"baseFees"`
	QuoteFees      float64 `json:"quoteFees"`
	LastPriceSwap  float64 `json:"lastPriceSwap"`
	LastPriceLiq   float64 `json:"lastPriceLiq"`
	LastPriceIndic float64 `json:"lastPriceIndic"`
	FeeRate        float64 `json:"feeRate"`
}

func (a *AccumPoolStats) Accumulate(e tables.AggEvent) {
	a.LatestTime = e.Time

	if e.IsFeeChange {
		a.accumFeeType(e)
	} else if e.IsSwap {
		a.accumSwapType(e)
	} else if e.IsLiq {
		a.accumLiqType(e)
	}
}

func (a *AccumPoolStats) accumFeeType(r tables.AggEvent) {
	FEE_RATE_MULTIPLIER := 10000.0 * 100
	a.FeeRate = float64(r.FeeRate) / FEE_RATE_MULTIPLIER
}

func (a *AccumPoolStats) accumLiqType(r tables.AggEvent) {
	a.accumulateFlows(r.BaseFlow, r.QuoteFlow)
	isStable := isFlowDualStable(r.BaseFlow, r.QuoteFlow)

	if isStable && r.FlowsAtMarket {
		if r.IsTickSkewed {
			price := derivePriceFromConcFlow(r.BaseFlow, r.QuoteFlow,
				r.BidTick, r.AskTick)
			if price != nil {
				a.LastPriceLiq = *price
				a.LastPriceIndic = *price
			}
		} else {
			a.LastPriceLiq = derivePriceFromAmbientFlow(r.BaseFlow, r.QuoteFlow)
			a.LastPriceIndic = a.LastPriceLiq
		}
	}
}

func (a *AccumPoolStats) accumSwapType(e tables.AggEvent) {
	a.accumulateFlows(e.BaseFlow, e.QuoteFlow)
	isStable := isFlowDualStable(e.BaseFlow, e.QuoteFlow)

	a.BaseVolume += math.Abs(e.BaseFlow)
	a.QuoteVolume += math.Abs(e.QuoteFlow)

	if e.InBaseQty {
		a.accumulateQuoteFees(e.QuoteFlow, a.FeeRate)
	} else {
		a.accumulateBaseFees(e.BaseFlow, a.FeeRate)
	}

	if isStable {
		price := derivePriceFromAmbientFlow(math.Abs(e.BaseFlow), math.Abs(e.QuoteFlow))
		a.LastPriceSwap = price
		a.LastPriceIndic = price
	}
}

func (a *AccumPoolStats) incrementFeeChange(r *tables.FeeChange) {
	a.FeeRate = float64(r.FeeRate) / 100 / 100
}

func (a *AccumPoolStats) accumulateFlows(baseFlow float64, quoteFlow float64) {
	a.BaseTvl += baseFlow
	a.QuoteTvl += quoteFlow
}

func (a *AccumPoolStats) accumulateBaseFees(baseFlow float64, feeRate float64) {
	a.BaseFees += math.Abs(baseFlow) * feeRate
}

func (a *AccumPoolStats) accumulateQuoteFees(quoteFlow float64, feeRate float64) {
	a.QuoteFees += math.Abs(quoteFlow) * feeRate
}
