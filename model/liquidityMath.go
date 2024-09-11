package model

import (
	"math"
	"math/big"

	"github.com/CrocSwap/graphcache-go/tables"
)

func deriveLiquidityFromAmbientFlow(baseFlow float64, quoteFlow float64) float64 {
	return math.Sqrt(baseFlow * quoteFlow)
}

func derivePriceFromAmbientFlow(baseFlow float64, quoteFlow float64) float64 {
	return math.Abs(baseFlow / quoteFlow)
}

func derivePriceFromSwapFlow(baseFlow float64, quoteFlow float64, feeRate float64, isBuy bool) float64 {
	if isBuy {
		return math.Abs(baseFlow/quoteFlow) * (1 + feeRate)
	} else {
		return math.Abs(baseFlow/quoteFlow) * (1 - feeRate)
	}
}

func DeriveLiquidityFromConcFlow(baseFlow float64, quoteFlow float64,
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
	solutionNeg := (-termB - math.Sqrt(termB*termB-4*termA*termC)) /
		(2 * termA)

	if solutionPos >= bidPrice && solutionPos <= askPrice {
		return solutionPos
	} else {
		return solutionNeg
	}
}

func DeriveTokensFromConcLiquidity(liquidity float64, bidTick int, askTick int, price float64) (baseTokens *big.Int, quoteTokens *big.Int) {
	if price == 0 {
		return nil, nil
	}
	bidPriceBig := big.NewFloat(tickToPrice(bidTick))
	askPriceBig := big.NewFloat(tickToPrice(askTick))
	liquidityBig := big.NewFloat(liquidity)
	clampedPriceBig := big.NewFloat(price)
	if big.NewFloat(price).Cmp(askPriceBig) == 1 {
		clampedPriceBig = askPriceBig
	} else if big.NewFloat(price).Cmp(bidPriceBig) == -1 {
		clampedPriceBig = bidPriceBig
	}
	sqrtClampedPriceBig := new(big.Float).Sqrt(clampedPriceBig)
	sqrtBidPriceBig := new(big.Float).Sqrt(bidPriceBig)
	sqrtAskPriceBig := new(big.Float).Sqrt(askPriceBig)

	baseTokensBig := new(big.Float).Sub(sqrtClampedPriceBig, sqrtBidPriceBig)
	baseTokensBig.Mul(baseTokensBig, liquidityBig)

	quoteTokensBig := new(big.Float).Sub(sqrtAskPriceBig, sqrtClampedPriceBig)
	quoteTokensBig.Quo(quoteTokensBig, new(big.Float).Mul(sqrtClampedPriceBig, sqrtAskPriceBig))
	quoteTokensBig.Mul(quoteTokensBig, liquidityBig)
	baseTokens, _ = baseTokensBig.Int(nil)
	quoteTokens, _ = quoteTokensBig.Int(nil)
	return
}

func DeriveTokensFromAmbLiquidity(liquidity float64, price float64) (baseTokens *big.Int, quoteTokens *big.Int) {
	if price == 0 {
		return nil, nil
	}
	price = math.Sqrt(price)
	baseTokens, _ = big.NewFloat(0).Mul(big.NewFloat(liquidity), big.NewFloat(price)).Int(nil)
	quoteTokens, _ = big.NewFloat(0).Quo(big.NewFloat(liquidity), big.NewFloat(price)).Int(nil)
	return
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

const MIN_NUMERIC_STABLE_FLOW = 1000

func isFlowNumericallyStable(baseFlow float64, quoteFlow float64) bool {
	return math.Abs(baseFlow) >= MIN_NUMERIC_STABLE_FLOW ||
		math.Abs(quoteFlow) >= MIN_NUMERIC_STABLE_FLOW
}

func isFlowDualStable(baseFlow float64, quoteFlow float64) bool {
	return math.Abs(baseFlow) >= MIN_NUMERIC_STABLE_FLOW &&
		math.Abs(quoteFlow) >= MIN_NUMERIC_STABLE_FLOW
}

func flowMagns(r *tables.LiqChange) (float64, float64) {
	return math.Abs(*r.BaseFlow), math.Abs(*r.QuoteFlow)
}
