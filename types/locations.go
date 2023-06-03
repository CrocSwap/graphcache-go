package types

type PoolLocation struct {
	ChainId ChainId    `json:"chainId"`
	Base    EthAddress `json:"base"`
	Quote   EthAddress `json:"quote"`
	PoolIdx int        `json:"poolIdx"`
}

type LiquidityLocation struct {
	BidTick int  `json:"bidTick"`
	AskTick int  `json:"askTick"`
	IsBid   bool `json:"isBid"` // Defaults to false if not valid
}

type PositionLocation struct {
	PoolLocation
	LiquidityLocation
	User EthAddress `json:"user"`
}

type KOClaimLocation struct {
	PositionLocation
	PivotTime int `json:"pivotTime"`
}

type BookLocation struct {
	PoolLocation
	LiquidityLocation
}

func AmbientLiquidityLocation() LiquidityLocation {
	return LiquidityLocation{BidTick: 0, AskTick: 0}
}

func RangeLiquidityLocation(bidTick int, askTick int) LiquidityLocation {
	return LiquidityLocation{BidTick: bidTick, AskTick: askTick}
}

func KnockoutRangeLocation(bidTick int, askTick int, isBid bool) LiquidityLocation {
	return LiquidityLocation{BidTick: bidTick, AskTick: askTick, IsBid: isBid}
}

func KnockoutTickLocation(tick int, isBid bool, tickWidth int) LiquidityLocation {
	if isBid {
		return LiquidityLocation{
			BidTick: tick,
			AskTick: tick + tickWidth,
			IsBid:   isBid,
		}
	} else {
		return LiquidityLocation{
			BidTick: tick - tickWidth,
			AskTick: tick,
			IsBid:   isBid,
		}
	}
}

func PositionTypeForLiq(loc LiquidityLocation) string {
	if loc.BidTick == 0 && loc.AskTick == 0 {
		return "ambient"
	} else {
		return "range"
	}
}

func (l PositionLocation) ToBookLoc() BookLocation {
	return BookLocation{
		l.PoolLocation,
		l.LiquidityLocation,
	}
}

func (l BookLocation) ToPositionLocation(user EthAddress) PositionLocation {
	return PositionLocation{
		l.PoolLocation,
		l.LiquidityLocation,
		user,
	}
}

func (l PositionLocation) ToClaimLoc(pivot int) KOClaimLocation {
	return KOClaimLocation{
		PositionLocation: l,
		PivotTime:        pivot,
	}
}

func (l BookLocation) ToClaimLoc(user EthAddress, pivotTime int) KOClaimLocation {
	return KOClaimLocation{
		l.ToPositionLocation(user),
		pivotTime,
	}
}

func (l LiquidityLocation) PivotTick() int {
	if l.IsBid {
		return l.BidTick
	} else {
		return l.AskTick
	}
}
