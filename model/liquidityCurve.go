package model

import (
	"log"

	"github.com/CrocSwap/graphcache-go/tables"
)

type LiquidityCurve struct {
	AmbientLiq float64
	Bumps      map[int]*LiquidityBump
}

type LiquidityBump struct {
	Tick             int     `json:"bumpTick"`
	LiquidityDelta   float64 `json:"liquidityDelta"`
	KnockoutBidLiq   float64 `json:"knockoutBidLiq"`
	KnockoutAskLiq   float64 `json:"knockoutAskLiq"`
	KnockoutBidWidth int     `json:"knockoutBidWidth"`
	KnockoutAskWidth int     `json:"knockoutAskWidth"`
	LatestUpdateTime int     `json:"latestUpdateTime"`
}

func NewLiquidityCurve() *LiquidityCurve {
	return &LiquidityCurve{
		AmbientLiq: 0,
		Bumps:      make(map[int]*LiquidityBump, 0),
	}
}

func (c *LiquidityCurve) UpdateLiqChange(l tables.LiqChange) {
	if l.PositionType == tables.PosTypeAmbient {
		liqMagn := determineLiquidityMagn(l)
		if l.ChangeType == tables.ChangeTypeBurn {
			liqMagn = -liqMagn
		}
		c.AmbientLiq += liqMagn
	}

	if l.ChangeType == tables.ChangeTypeMint || l.ChangeType == tables.ChangeTypeBurn {
		c.updateUserLiq(l)
	}

	if l.ChangeType == tables.ChangeTypeCross {
		c.updateKOCross(l)
	}
}

func (c *LiquidityCurve) updateUserLiq(l tables.LiqChange) {
	bidBump := c.materializeBump(l.BidTick)
	askBump := c.materializeBump(l.AskTick)

	liqMagn := determineLiquidityMagn(l)
	if l.ChangeType == tables.ChangeTypeBurn {
		liqMagn = -liqMagn
	}

	bidBump.IncrLiquidity(liqMagn, l.Time)
	askBump.IncrLiquidity(-liqMagn, l.Time)

	if l.PositionType == tables.PosTypeKnockout {
		if l.IsBid > 0 {
			bidBump.IncrKOBid(liqMagn, l.AskTick)
		} else {
			askBump.IncrKOAsk(-liqMagn, l.BidTick)
		}
	}
}

func (c *LiquidityCurve) updateKOCross(k tables.LiqChange) {
	if k.IsBid > 0 {
		bidBump := c.materializeBump(k.BidTick)
		koLiq, joinTick := bidBump.KnockoutBid(k.Time)

		askBump := c.materializeBump(joinTick)
		askBump.IncrLiquidity(koLiq, k.Time)

	} else {
		askBump := c.materializeBump(k.AskTick)
		koLiq, joinTick := askBump.KnockoutAsk(k.Time)

		bidBump := c.materializeBump(joinTick)
		bidBump.IncrLiquidity(koLiq, k.Time)
	}
}

func (b *LiquidityBump) IncrLiquidity(liqDelta float64, time int) {
	b.updateTime(time)
	b.LiquidityDelta += liqDelta
}

func (b *LiquidityBump) IncrKOBid(liqDelta float64, joinTick int) {
	b.KnockoutBidLiq += liqDelta
	b.KnockoutBidWidth = joinTick - b.Tick
}

func (b *LiquidityBump) IncrKOAsk(liqDelta float64, joinTick int) {
	b.KnockoutAskLiq += liqDelta
	b.KnockoutAskWidth = b.Tick - joinTick
}

func (b *LiquidityBump) KnockoutBid(time int) (float64, int) {
	b.updateTime(time)
	koLiq := b.KnockoutBidLiq
	b.LiquidityDelta -= koLiq

	joinTick := b.Tick + b.KnockoutBidWidth
	b.KnockoutBidWidth = 0

	b.KnockoutBidLiq = 0
	return koLiq, joinTick
}

func (b *LiquidityBump) KnockoutAsk(time int) (float64, int) {
	b.updateTime(time)
	koLiq := b.KnockoutAskLiq
	b.LiquidityDelta -= koLiq

	joinTick := b.Tick - b.KnockoutAskWidth
	b.KnockoutAskWidth = 0

	b.KnockoutAskLiq = 0
	return koLiq, joinTick
}

func (b *LiquidityBump) updateTime(time int) {
	if time < b.LatestUpdateTime {
		log.Fatalf("Liquidity curve updated out of time order %d -> %d",
			b.LatestUpdateTime, time)
	}
	b.LatestUpdateTime = time
}

func (c *LiquidityCurve) materializeBump(tick int) *LiquidityBump {
	lookup, ok := c.Bumps[tick]
	if !ok {
		lookup = &LiquidityBump{
			Tick: tick,
		}
		c.Bumps[tick] = lookup
	}
	return lookup
}

func determineLiquidityMagn(r tables.LiqChange) float64 {
	baseFlow, quoteFlow := flowMagns(&r)

	if !isFlowNumericallyStable(baseFlow, quoteFlow) {
		return 0
	}

	if r.PositionType == tables.PosTypeAmbient {
		return deriveLiquidityFromAmbientFlow(baseFlow, quoteFlow)
	} else {
		return DeriveLiquidityFromConcFlow(baseFlow, quoteFlow, r.BidTick, r.AskTick)
	}
}
