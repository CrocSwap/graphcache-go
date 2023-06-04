package model

import (
	"log"

	"github.com/CrocSwap/graphcache-go/tables"
)

type LiquidityDeltaHist struct {
	hist []LiquidityDelta
}

type LiquidityDelta struct {
	time         int
	liqChange    float64
	resetRewards bool
}

func (l *LiquidityDeltaHist) weightedAverageTime() int {
	openLiq := 0.0
	openTime := 0.0

	for _, delta := range l.hist {
		if delta.resetRewards == true {
			openTime = float64(delta.time)
		}

		if delta.liqChange < 0 {
			openLiq = openLiq + delta.liqChange
			if openLiq < MIN_NUMERIC_STABLE_FLOW {
				openLiq = 0
			}
		}

		if delta.liqChange > 0 {
			weight := openLiq / (openLiq + delta.liqChange)
			openTime = openTime*weight + float64(delta.time)*(1.0-weight)
		}
	}
	return int(openTime)
}

func (l *LiquidityDeltaHist) appendChange(r tables.LiqChange) {
	l.initHist()
	l.assertTimeForward(r.Time)

	if r.ChangeType == "harvest" {
		l.hist = append(l.hist, LiquidityDelta{
			time:         r.Time,
			resetRewards: true,
		})

	} else {
		liqMagn := l.determineLiquidityMagn(r)

		if r.ChangeType == "mint" {
			l.hist = append(l.hist, LiquidityDelta{
				time:      r.Time,
				liqChange: -liqMagn})
		} else if r.ChangeType == "burn" {
			l.hist = append(l.hist, LiquidityDelta{
				time:      r.Time,
				liqChange: -liqMagn})
		}
	}
}

func (l *LiquidityDeltaHist) determineLiquidityMagn(r tables.LiqChange) float64 {
	if !isFlowNumericallyStable(*r.BaseFlow, *r.QuoteFlow) {
		return 0
	}

	if r.PositionType == "ambient" {
		return deriveLiquidityFromAmbientFlow(*r.BaseFlow, *r.QuoteFlow)
	} else {
		return deriveLiquidityFromConcFlow(*r.BaseFlow, *r.QuoteFlow, r.BidTick, r.AskTick)
	}
}

func (l *LiquidityDeltaHist) initHist() {
	if l.hist == nil {
		l.hist = make([]LiquidityDelta, 0)
	}
}

func (l *LiquidityDeltaHist) assertTimeForward(time int) {
	if len(l.hist) == 0 {
		return
	}
	lastTime := l.hist[0].time

	if time < lastTime {
		log.Fatalf("Liquidity delta history has backward time step %d->%d", lastTime, time)
	}
}
