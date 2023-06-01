package types

type PoolLocation struct {
	ChainId ChainId    `json:"chainId"`
	Base    EthAddress `json:"base"`
	Quote   EthAddress `json:"quote"`
	PoolIdx int        `json:"poolIdx"`
}

type LiquidityLocation struct {
	BidTick   int  `json:"bidTick"`
	AskTick   int  `json:"askTick"`
	PivotTime int  `json:"pivotTime"`
	IsBid     bool `json:"isBid"`
}

type PositionLocation struct {
	PoolLocation
	LiquidityLocation
	User EthAddress `json:"user"`
}

func AmbientLiquidityLocation() LiquidityLocation {
	return LiquidityLocation{BidTick: 0, AskTick: 0, PivotTime: 0, IsBid: false}
}

func RangeLiquidityLocation(bidTick int, askTick int) LiquidityLocation {
	return LiquidityLocation{BidTick: bidTick, AskTick: askTick, PivotTime: 0, IsBid: false}
}

func KnockoutLiquidityLocation(bidTick int, askTick int, pivotTime int, isBid bool) LiquidityLocation {
	return LiquidityLocation{BidTick: bidTick, AskTick: askTick,
		PivotTime: pivotTime, IsBid: isBid}
}

func PositionTypeForLiq(loc LiquidityLocation) string {
	if loc.BidTick == 0 && loc.AskTick == 0 {
		return "ambient"
	} else if loc.PivotTime == 0 {
		return "range"
	} else {
		return "knockout"
	}
}
