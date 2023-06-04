package model

import "math"

func deriveLiquidityFromAmbientFlow(baseFlow float64, quoteFlow float64) float64 {
	return math.Sqrt(baseFlow * quoteFlow)
}

func derivePriceFromAmbientFlow(baseFlow float64, quoteFlow float64) float64 {
	return baseFlow / quoteFlow
}

func derivePriceFromSwapFlow(baseFlow float64, quoteFlow float64) float64 {
	return math.Abs(baseFlow / quoteFlow)
}

func deriveLiquidityFromConcFlow(baseFlow float64, quoteFlow float64,
	bidTick int, askTick int) float64 {
	bidPrice := math.Sqrt(tickToPrice(bidTick))
	askPrice := math.Sqrt(tickToPrice(askTick))

	if quoteFlow == 0 {
		return baseFlow / (askPrice - bidPrice)
	} else if baseFlow == 0 {
		return quoteFlow / (1/askPrice - 1/bidPrice)
	} else {
		price := *derivePriceFromConcFlow(baseFlow, quoteFlow, bidTick, askTick)
		return baseFlow / (price - bidPrice)
	}
}

func derivePriceFromConcFlow(baseFlow float64, quoteFlow float64,
	bidTick int, askTick int) *float64 {
	if quoteFlow == 0 {
		return nil
	} else if baseFlow == 0 {
		return nil
	} else {
		price := derivePriceFromInRange(baseFlow, quoteFlow, bidTick, askTick)
		return &price
	}
}

func derivePriceFromInRange(baseFlow float64, quoteFlow float64,
	bidTick int, askTick int) float64 {
	bidPrice := math.Sqrt(tickToPrice(bidTick))
	askPrice := math.Sqrt(tickToPrice(askTick))

	termA := quoteFlow * askPrice
	termB := baseFlow - quoteFlow*bidPrice*askPrice
	termC := -baseFlow * askPrice

	solutionPos := (-termB + math.Sqrt(termB*termB-4*termA*termC)) /
		(2 * termA)
	solutionNeg := (-termB + math.Sqrt(termB*termB-4*termA*termC)) /
		(2 * termA)

	if solutionPos >= bidPrice && solutionPos <= askPrice {
		return solutionPos
	} else {
		return solutionNeg
	}
}

func tickToPrice(tick int) float64 {
	return math.Pow(1.0001, float64(tick))
}

const MIN_NUMERIC_STABLE_FLOW = 10000

func isFlowNumericallyStable(baseFlow float64, quoteFlow float64) bool {
	return baseFlow > MIN_NUMERIC_STABLE_FLOW &&
		quoteFlow > MIN_NUMERIC_STABLE_FLOW
}
