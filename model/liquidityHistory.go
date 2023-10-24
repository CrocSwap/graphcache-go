package model

import (
	"log"
	"time"

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

func (l *LiquidityDeltaHist) netCumulativeLiquidity() float64 {
	totalLiq := 0.0

	for _, delta := range l.hist {
		totalLiq += delta.liqChange
	}

	if totalLiq < getMinNumericStableFlow() {
		return 0
	}
	return totalLiq
}

func (l *LiquidityDeltaHist) weightedAverageDuration() float64 {
	present := float64(time.Now().Unix())
	past := float64(l.weightedAverageTime())
	return present - past
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
			if openLiq < 0 || openLiq < getMinNumericStableFlow() {
				openLiq = 0
			}
		}

		if delta.liqChange > 0 {
			weight := openLiq / (openLiq + delta.liqChange)
			openTime = openTime*weight + float64(delta.time)*(1.0-weight)
		}

		if delta.liqChange == 0 && openLiq == 0 {
			openTime = float64(delta.time)
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
		liqMagn := determineLiquidityMagn(r)

		if r.ChangeType == "mint" {
			l.hist = append(l.hist, LiquidityDelta{
				time:      r.Time,
				liqChange: liqMagn})

		} else if r.ChangeType == "burn" {
			l.hist = append(l.hist, LiquidityDelta{
				time:      r.Time,
				liqChange: -liqMagn})
		}
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
