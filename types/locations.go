package types

type PoolLocation struct {
	ChainId ChainId    `json:"chainId"`
	Base    EthAddress `json:"base"`
	Quote   EthAddress `json:"quote"`
	PoolIdx int        `json:"poolIdx"`
}

type LiquidityLocation struct {
	BidTick int `json:"bidTick"`
	AskTick int `json:"askTick"`
}

type PositionLocation struct {
	PoolLocation
	LiquidityLocation
	User EthAddress `json:"user"`
}

func AmbientLiquidityLocation() LiquidityLocation {
	return LiquidityLocation{BidTick: 0, AskTick: 0}
}

func RangeLiquidityLocation(bidTick int, askTick int) LiquidityLocation {
	return LiquidityLocation{BidTick: bidTick, AskTick: askTick}
}

func PositionTypeForLiq(loc LiquidityLocation) string {
	if loc.BidTick == 0 && loc.AskTick == 0 {
		return "ambient"
	} else {
		return "range"
	}
}
