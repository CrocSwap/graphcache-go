package model

import (
	"math"

	"github.com/CrocSwap/graphcache-go/utils"
)

type CandleBuilder struct {
	series      []Candle
	running     RunningCandle
	period      int
	atValidHist bool
}

type RunningCandle struct {
	candle          Candle
	lastAccum       AccumPoolStats
	openCumBaseVol  float64
	openCumQuoteVol float64

	accumPoolStats []AccumPoolStats


}

type Candle struct {
	PriceOpen    float64 `json:"priceOpen"`
	PriceClose   float64 `json:"priceClose"`
	MinPrice     float64 `json:"minPrice"`
	MaxPrice     float64 `json:"maxPrice"`

	VolumeBase   float64 `json:"volumeBase"`
	VolumeQuote  float64 `json:"volumeQuote"`
	TvlBase      float64 `json:"tvlBase"`
	TvlQuote     float64 `json:"tvlQuote"`
	FeeRateOpen  float64 `json:"feeRateOpen"`
	FeeRateClose float64 `json:"feeRateClose"`
	
	Period       int     `json:"period"`
	Time         int     `json:"time"`

	IsDecimalized bool   `json:"isDecimalized"`



}

var uniswapCandles = utils.GoDotEnvVariable("UNISWAP_CANDLES") == "true"
var MevThreshold = utils.GetEnvVarIntFromString("MEV_THRESHOLD", 1000000)
var EnableStdDevFilter = utils.GoDotEnvVariable("ENABLE_MAD_FILTER") == "true"
func NewCandleBuilder(startTime int, period int, open AccumPoolStats) *CandleBuilder {
	builder := &CandleBuilder{
		series:      make([]Candle, 0),
		running:     RunningCandle{},
		period:      period,
		atValidHist: false,
	}
	builder.openCandle(open, startTime)
	return builder
}

func (c *CandleBuilder) openCandle(accum AccumPoolStats, startTime int) {
	c.running.candle = Candle{
		PriceOpen:    accum.LastPriceSwap,
		PriceClose:   accum.LastPriceSwap,
		MinPrice:     accum.LastPriceSwap,
		MaxPrice:     accum.LastPriceSwap,
		VolumeBase:   0.0,
		VolumeQuote:  0.0,
		TvlBase:      accum.BaseTvl,
		TvlQuote:     accum.QuoteTvl,
		FeeRateOpen:  accum.FeeRate,
		FeeRateClose: accum.FeeRate,
		Period:       c.period,
		Time:         startTime,
		IsDecimalized: uniswapCandles,
	}

	c.running.lastAccum = accum
	c.running.openCumBaseVol = accum.BaseVolume
	c.running.openCumQuoteVol = accum.QuoteVolume
	c.running.accumPoolStats =  make([]AccumPoolStats, 0)

}

func (c *CandleBuilder) Close(endTime int) []Candle {
	// Question 6: Should this be an if statement?
	for c.running.candle.Time+c.period < endTime {
		c.closeCandle()
	}
	return c.series
}
func getMinValidLiqudity() float64 {
	uniswapCandles := utils.GoDotEnvVariable("UNISWAP_CANDLES") == "true"
	if(uniswapCandles){
		return 1
	} else {
		return 100000
	}
}

func (c *CandleBuilder) closeCandle() {
	MIN_VALID_LIQUIDITY := getMinValidLiqudity()
	
	c.atValidHist = c.atValidHist ||
		c.running.candle.TvlBase >= MIN_VALID_LIQUIDITY ||
		c.running.candle.TvlQuote >= MIN_VALID_LIQUIDITY

	if c.atValidHist {
		c.series = append(c.series, c.running.candle)
	}

	c.openCandle(c.running.lastAccum, c.running.candle.Time+c.period)
}



func (c *CandleBuilder) accumulateCandle(rollingStdDev RollingStdDev) {
	valid := 0
	filtered := 0

	for _, accum := range c.running.accumPoolStats {

		if(!EnableStdDevFilter || AllowedThroughStdDevFilter(accum, rollingStdDev, c.running.candle.PriceOpen, c.running.candle.PriceClose)){
			valid += 1
			if accum.LastPriceSwap < c.running.candle.MinPrice {
				c.running.candle.MinPrice = accum.LastPriceSwap
			}
			if accum.LastPriceSwap > c.running.candle.MaxPrice {
				c.running.candle.MaxPrice = accum.LastPriceSwap
			}

			c.running.candle.VolumeBase = accum.BaseVolume - c.running.openCumBaseVol
			c.running.candle.VolumeQuote = accum.QuoteVolume - c.running.openCumQuoteVol

			c.running.candle.TvlBase = accum.BaseTvl
			c.running.candle.TvlQuote = accum.QuoteTvl

			c.running.candle.FeeRateClose = accum.FeeRate
		} else {
			filtered += 1
		}
	} 

	c.running.accumPoolStats = make([]AccumPoolStats, 0)
	c.closeCandle()
}

func (c *CandleBuilder) Increment(accum AccumPoolStats, rollingStdDev RollingStdDev)  {
	for accum.LatestTime >= c.running.candle.Time+c.period {
		c.accumulateCandle(rollingStdDev)
	}
	c.running.candle.PriceClose = accum.LastPriceSwap
	c.running.accumPoolStats = append(c.running.accumPoolStats, accum)
	c.running.lastAccum = accum
} 

func AllowedThroughStdDevFilter(accum AccumPoolStats, rollingStdDev RollingStdDev, PriceOpen float64, PriceClose float64) bool {
	thresholdMin := math.Min(PriceOpen, PriceClose)
	thresholdMax := math.Max(PriceOpen, PriceClose)
	stdev := rollingStdDev[accum.LatestTime]
	var mevMargin = 0.0
	if accum.LastPriceSwap < thresholdMin {
		mevMargin = (thresholdMin - accum.LastPriceSwap) / stdev
	} else if accum.LastPriceSwap > thresholdMax {
		mevMargin = (accum.LastPriceSwap - thresholdMax )/ stdev
	}


	if mevMargin > float64(MevThreshold){
		return false
	}
	

	return true
}

func calculateMEVMargin(Price float64, stdev, thresholdMin float64, thresholdMax float64) float64{
	var mevMargin = 0.0
	if Price < thresholdMin {
		mevMargin = (thresholdMin - Price) / stdev
	} else if Price > thresholdMax {
		mevMargin = (Price - thresholdMax )/ stdev
	}
	return mevMargin
}