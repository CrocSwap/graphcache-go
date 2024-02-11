package types

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
)

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

func (p PositionLocation) Hash() [32]byte {
	buf := new(bytes.Buffer)
	buf.Grow(160)
	buf.WriteString(string(p.ChainId))
	buf.WriteString(string(p.Base))
	buf.WriteString(string(p.Quote))
	binary.Write(buf, binary.BigEndian, int32(p.PoolIdx))
	binary.Write(buf, binary.BigEndian, int32(p.BidTick))
	binary.Write(buf, binary.BigEndian, int32(p.AskTick))
	binary.Write(buf, binary.BigEndian, p.IsBid)
	buf.WriteString(string(p.User))
	return sha256.Sum256(buf.Bytes())
}

func (k KOClaimLocation) Hash() [32]byte {
	buf := new(bytes.Buffer)
	buf.Grow(160)
	buf.WriteString(string(k.ChainId))
	buf.WriteString(string(k.Base))
	buf.WriteString(string(k.Quote))
	binary.Write(buf, binary.BigEndian, int32(k.PoolIdx))
	binary.Write(buf, binary.BigEndian, int32(k.BidTick))
	binary.Write(buf, binary.BigEndian, int32(k.AskTick))
	binary.Write(buf, binary.BigEndian, k.IsBid)
	buf.WriteString(string(k.User))
	binary.Write(buf, binary.BigEndian, int32(k.PivotTime))
	return sha256.Sum256(buf.Bytes())
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
