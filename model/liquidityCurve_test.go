package model

import (
	"math"
	"testing"

	"github.com/CrocSwap/graphcache-go/tables"
)

func TestAmbientNoop(t *testing.T) {
	curve := NewLiquidityCurve()
	liqFlow := 30000.0
	curve.UpdateLiqChange(tables.LiqChange{
		BidTick:      -250,
		AskTick:      500,
		ChangeType:   "mint",
		PositionType: "ambient",
		BaseFlow:     &liqFlow,
		QuoteFlow:    &liqFlow,
		IsBid:        1,
	})
	if len(curve.Bumps) > 0 {
		t.Fatal("Ambient update created tick bump")
	}
}

func TestRangeMint(t *testing.T) {
	curve := NewLiquidityCurve()
	liqFlow := 30000.0
	curve.UpdateLiqChange(tables.LiqChange{
		BidTick:      -250,
		AskTick:      500,
		ChangeType:   "mint",
		PositionType: "concentrated",
		BaseFlow:     &liqFlow,
		QuoteFlow:    &liqFlow,
	})

	bidLiq := curve.Bumps[-250].LiquidityDelta
	askLiq := curve.Bumps[500].LiquidityDelta

	if bidLiq <= 0 {
		t.Fatalf("Lower liquidity range not positive %f", bidLiq)
	}
	if askLiq >= 0 {
		t.Fatalf("Upper liquidity range not negative %f", bidLiq)
	}
	if bidLiq != -askLiq {
		t.Fatalf("Liquidity range mismatch %f <-> %f", bidLiq, askLiq)
	}
}

func TestKnockoutBid(t *testing.T) {
	curve := NewLiquidityCurve()

	liqFlow := 30000.0
	curve.UpdateLiqChange(tables.LiqChange{
		BidTick:      -250,
		AskTick:      500,
		ChangeType:   "mint",
		PositionType: "concentrated",
		BaseFlow:     &liqFlow,
		QuoteFlow:    &liqFlow,
	})

	startBidLiq := curve.Bumps[-250].LiquidityDelta
	startAskLiq := curve.Bumps[500].LiquidityDelta

	liqFlow = 250000.0
	curve.UpdateLiqChange(tables.LiqChange{
		BidTick:      -250,
		AskTick:      500,
		ChangeType:   "mint",
		PositionType: "knockout",
		BaseFlow:     &liqFlow,
		QuoteFlow:    &liqFlow,
		IsBid:        1,
	})

	koCrossRow := tables.LiqChange{
		BidTick:      -250,
		AskTick:      -250,
		ChangeType:   "cross",
		PositionType: "knockout",
	}
	curve.UpdateLiqChange(koCrossRow)

	bidBump := curve.Bumps[-250]
	if math.Abs(bidBump.LiquidityDelta-startBidLiq) > 0.0001 {
		t.Fatalf(`Mismatched bid liq %f (expected %f)`, bidBump.LiquidityDelta, startBidLiq)
	}
	if bidBump.KnockoutBidLiq != 0 {
		t.Fatalf(`Mismatched bid ko liq %f (expected %f)`, bidBump.KnockoutBidLiq, 0.0)
	}

	askBump := curve.Bumps[500]
	if math.Abs(askBump.LiquidityDelta-startAskLiq) > 0.0001 {
		t.Fatalf(`Mismatched ask liq %f (expected %f)`, askBump.LiquidityDelta, startAskLiq)
	}
}
