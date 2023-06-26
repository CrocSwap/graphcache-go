package model

import (
	"math"

	"github.com/CrocSwap/graphcache-go/tables"
	"github.com/CrocSwap/graphcache-go/utils"
)

func deriveLiquidityFromAmbientFlow(baseFlow float64, quoteFlow float64) float64 {
	return math.Sqrt(baseFlow * quoteFlow)
}

func derivePriceFromAmbientFlow(baseFlow float64, quoteFlow float64) float64 {
	return math.Abs(baseFlow / quoteFlow)
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
		return quoteFlow / (1/bidPrice - 1/askPrice)
	} else {
		price := *deriveRootPriceFromConcFlow(baseFlow, quoteFlow, bidTick, askTick)
		return baseFlow / (price - bidPrice)
	}
}

func derivePriceFromConcFlow(baseFlow float64, quoteFlow float64,
	bidTick int, askTick int) *float64 {
	root := deriveRootPriceFromConcFlow(baseFlow, quoteFlow, bidTick, askTick)
	if root == nil {
		return nil
	}
	price := *root * *root
	return &price
}

func deriveRootPriceFromConcFlow(baseFlow float64, quoteFlow float64,
	bidTick int, askTick int) *float64 {
	if quoteFlow == 0 {
		return nil
	} else if baseFlow == 0 {
		return nil
	} else {
		price := deriveRootFromInRange(baseFlow, quoteFlow, bidTick, askTick)
		return &price
	}
}

func deriveRootFromInRange(baseFlow float64, quoteFlow float64,
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

func estLiqAmplification(bidTick int, askTick int) float64 {
	midTick := (bidTick + askTick) / 2
	bidPrice := math.Sqrt(tickToPrice(bidTick))
	midPrice := math.Sqrt(tickToPrice(midTick))
	return midPrice / (midPrice - bidPrice)
}

func tickToPrice(tick int) float64 {
	return math.Pow(1.0001, float64(tick))
}

func getMinNumericStableFlow() float64 {
	uniswapCandles := utils.GoDotEnvVariable("UNISWAP_CANDLES") == "true"
	if(uniswapCandles){
		return 0.01
	} else {
		return 1000
	}
}




func isFlowNumericallyStable(baseFlow float64, quoteFlow float64) bool {
	return math.Abs(baseFlow) >= getMinNumericStableFlow() ||
		math.Abs(quoteFlow) >= getMinNumericStableFlow()
}

func isFlowDualStable(baseFlow float64, quoteFlow float64) bool {
	return math.Abs(baseFlow) > getMinNumericStableFlow() &&
		math.Abs(quoteFlow) > getMinNumericStableFlow()
}

func flowMagns(r *tables.LiqChange) (float64, float64) {
	return math.Abs(*r.BaseFlow), math.Abs(*r.QuoteFlow)
}
