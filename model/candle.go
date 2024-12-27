package model

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
}

func NewCandleBuilder(startTime int, period int, open AccumPoolStats) *CandleBuilder {
	builder := &CandleBuilder{
		series:      make([]Candle, 0, 1),
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
	}

	c.running.lastAccum = accum
	c.running.openCumBaseVol = accum.BaseVolume
	c.running.openCumQuoteVol = accum.QuoteVolume
}

func (c *CandleBuilder) Close(endTime int) []Candle {
	for c.running.candle.Time+c.period < endTime {
		c.closeCandle()
	}
	return c.series
}

func (c *CandleBuilder) closeCandle() {
	MIN_VALID_LIQUIDITY := 100000.0

	c.atValidHist = c.atValidHist ||
		c.running.candle.TvlBase >= MIN_VALID_LIQUIDITY ||
		c.running.candle.TvlQuote >= MIN_VALID_LIQUIDITY
	// c.atValidHist = true

	if c.atValidHist {
		c.series = append(c.series, c.running.candle)
	}

	c.openCandle(c.running.lastAccum, c.running.candle.Time+c.period)
}

func (c *CandleBuilder) Increment(accum AccumPoolStats) {
	for accum.LatestTime >= c.running.candle.Time+c.period {
		c.closeCandle()
	}

	c.running.candle.PriceClose = accum.LastPriceSwap
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
	c.running.lastAccum = accum
}

// Given a list of hourly candles, combine them into a list of `n` candles with
// `hours*3600` period from `startTime` to `endTime`.
func CombineHourlyCandles(candles []Candle, hours int, startTime int, endTime int, n int) []Candle {
	combined := make([]Candle, 0, max((endTime-startTime)/hours/3600, 1)) // max(1, ...) so that JSON is always an array
	period := hours * 3600
	startGroup := -1
	if len(candles) > 0 {
		// log.Println("got candles:", candles[0].Time, candles[len(candles)-1].Time)
	}
	for _, candle := range candles {
		if candle.Time < startTime {
			// log.Println("Skipping candle", candle.Time, startTime)
			continue
		}
		if candle.Time >= endTime || len(combined) >= n {
			// log.Println("Breaking at", candle.Time, endTime, len(combined), n)
			break
		}
		if startGroup == -1 {
			startGroup = (candle.Time - (candle.Time % period)) / period
		}
		ci := (candle.Time-(candle.Time%period))/period - startGroup

		// log.Println("Processing candle", candle.Time, startTime, endTime, ci)
		if ci >= len(combined) {
			combined = append(combined, Candle{})
		}
		combCandle := &combined[ci]

		if combCandle.Time == 0 {
			combCandle.Time = candle.Time - candle.Time%period
			combCandle.Period = period
			combCandle.PriceOpen = candle.PriceOpen
			combCandle.MinPrice = candle.MinPrice
			combCandle.MaxPrice = candle.MaxPrice
			combCandle.FeeRateOpen = candle.FeeRateOpen
		}

		combCandle.PriceClose = candle.PriceClose
		combCandle.VolumeBase += candle.VolumeBase
		combCandle.VolumeQuote += candle.VolumeQuote
		combCandle.TvlBase = candle.TvlBase
		combCandle.TvlQuote = candle.TvlQuote
		combCandle.FeeRateClose = candle.FeeRateClose
		if combCandle.MaxPrice < candle.MaxPrice {
			combCandle.MaxPrice = candle.MaxPrice
		}
		if combCandle.MinPrice > candle.MinPrice {
			combCandle.MinPrice = candle.MinPrice
		}

	}
	return combined
}
