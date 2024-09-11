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
		ChangeType:   tables.ChangeTypeMint,
		PositionType: tables.PosTypeAmbient,
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
		ChangeType:   tables.ChangeTypeMint,
		PositionType: tables.PosTypeConcentrated,
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
		ChangeType:   tables.ChangeTypeMint,
		PositionType: tables.PosTypeConcentrated,
		BaseFlow:     &liqFlow,
		QuoteFlow:    &liqFlow,
	})

	startBidLiq := curve.Bumps[-250].LiquidityDelta
	startAskLiq := curve.Bumps[500].LiquidityDelta

	liqFlow = 250000.0
	curve.UpdateLiqChange(tables.LiqChange{
		BidTick:      -250,
		AskTick:      500,
		ChangeType:   tables.ChangeTypeMint,
		PositionType: tables.PosTypeKnockout,
		BaseFlow:     &liqFlow,
		QuoteFlow:    &liqFlow,
		IsBid:        1,
	})

	curve.UpdateLiqChange(tables.LiqChange{
		BidTick:      -250,
		AskTick:      -250,
		ChangeType:   tables.ChangeTypeCross,
		PositionType: tables.PosTypeKnockout,
		IsBid:        1,
	})

	bidBump := curve.Bumps[-250]
	if math.Abs(bidBump.LiquidityDelta-startBidLiq) > 0.0001 {
		t.Fatalf(`Mismatched bid liq %f (expected %f)`, bidBump.LiquidityDelta, startBidLiq)
	}
	if bidBump.KnockoutBidLiq != 0 {
		t.Fatalf(`Mismatched bid ko liq %f (expected %f)`, bidBump.KnockoutBidLiq, 0.0)
	}
	if bidBump.KnockoutBidWidth != 0 {
		t.Fatalf("Knockout bid width not reset")
	}

	askBump := curve.Bumps[500]
	if math.Abs(askBump.LiquidityDelta-startAskLiq) > 0.0001 {
		t.Fatalf(`Mismatched ask liq %f (expected %f)`, askBump.LiquidityDelta, startAskLiq)
	}
}

func TestKnockoutAsk(t *testing.T) {
	curve := NewLiquidityCurve()

	liqFlow := 30000.0
	curve.UpdateLiqChange(tables.LiqChange{
		BidTick:      -250,
		AskTick:      500,
		ChangeType:   tables.ChangeTypeMint,
		PositionType: tables.PosTypeConcentrated,
		BaseFlow:     &liqFlow,
		QuoteFlow:    &liqFlow,
	})

	startBidLiq := curve.Bumps[-250].LiquidityDelta
	startAskLiq := curve.Bumps[500].LiquidityDelta

	liqFlow = 250000.0
	curve.UpdateLiqChange(tables.LiqChange{
		BidTick:      -250,
		AskTick:      500,
		ChangeType:   tables.ChangeTypeMint,
		PositionType: tables.PosTypeKnockout,
		BaseFlow:     &liqFlow,
		QuoteFlow:    &liqFlow,
		IsBid:        0,
	})

	curve.UpdateLiqChange(tables.LiqChange{
		BidTick:      500,
		AskTick:      500,
		ChangeType:   tables.ChangeTypeCross,
		PositionType: tables.PosTypeKnockout,
		IsBid:        0,
	})

	askBump := curve.Bumps[500]
	if math.Abs(askBump.LiquidityDelta-startAskLiq) > 0.0001 {
		t.Fatalf(`Mismatched ask liq %f (expected %f)`, askBump.LiquidityDelta, startAskLiq)
	}
	if askBump.KnockoutAskLiq != 0 {
		t.Fatalf(`Mismatched ask ko liq %f (expected %f)`, askBump.KnockoutBidLiq, 0.0)
	}
	if askBump.KnockoutAskWidth != 0 {
		t.Fatalf("Knockout ask width not reset")
	}

	bidBump := curve.Bumps[-250]
	if math.Abs(askBump.LiquidityDelta-startAskLiq) > 0.0001 {
		t.Fatalf(`Mismatched ask liq %f (expected %f)`, bidBump.LiquidityDelta, startBidLiq)
	}
}
