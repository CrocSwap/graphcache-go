package model

import (
	"log"
	"time"

	"github.com/CrocSwap/graphcache-go/tables"
)

type LiquidityDeltaHist struct {
	Hist []LiquidityDelta `json:"hist"`
}

type LiquidityDelta struct {
	Time         int
	LiqChange    float64
	resetRewards bool
}

func (l *LiquidityDeltaHist) netCumulativeLiquidity() float64 {
	totalLiq := 0.0

	for _, delta := range l.Hist {
		totalLiq += delta.LiqChange
	}

	if totalLiq < MIN_NUMERIC_STABLE_FLOW {
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

	for _, delta := range l.Hist {
		if delta.resetRewards == true {
			openTime = float64(delta.Time)
		}

		if delta.LiqChange < 0 {
			openLiq = openLiq + delta.LiqChange
			if openLiq < 0 || openLiq < MIN_NUMERIC_STABLE_FLOW {
				openLiq = 0
			}
		}

		if delta.LiqChange > 0 {
			weight := openLiq / (openLiq + delta.LiqChange)
			openTime = openTime*weight + float64(delta.Time)*(1.0-weight)
		}

		if delta.LiqChange == 0 && openLiq == 0 {
			openTime = float64(delta.Time)
		}
	}
	return int(openTime)
}

func (l *LiquidityDeltaHist) appendChange(r tables.LiqChange) {
	l.initHist()
	l.assertTimeForward(r.Time)

	if r.ChangeType == tables.ChangeTypeHarvest {
		l.Hist = append(l.Hist, LiquidityDelta{
			Time:         r.Time,
			resetRewards: true,
		})

	} else {
		liqMagn := determineLiquidityMagn(r)

		if r.ChangeType == tables.ChangeTypeMint {
			l.Hist = append(l.Hist, LiquidityDelta{
				Time:      r.Time,
				LiqChange: liqMagn})

		} else if r.ChangeType == tables.ChangeTypeBurn {
			l.Hist = append(l.Hist, LiquidityDelta{
				Time:      r.Time,
				LiqChange: -liqMagn})
		}
	}
}

func (l *LiquidityDeltaHist) initHist() {
	if l.Hist == nil {
		l.Hist = make([]LiquidityDelta, 0)
	}
}

func (l *LiquidityDeltaHist) assertTimeForward(time int) {
	if len(l.Hist) == 0 {
		return
	}
	lastTime := l.Hist[0].Time

	if time < lastTime {
		log.Fatalf("Liquidity delta history has backward time step %d->%d", lastTime, time)
	}
}
