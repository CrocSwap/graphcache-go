package types

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"

	"github.com/CrocSwap/graphcache-go/tables"
)

type PoolTxEvent struct {
	EthTxHeader
	PoolLocation
	PoolEventFlow
	PoolEventDescriptor
	PoolRangeFields
}

type EthTxHeader struct {
	BlockNum  int        `json:"blockNum"`
	TxHash    EthTxHash  `json:"txHash"`
	TxTime    int        `json:"txTime"`
	User      EthAddress `json:"user"`
	CallIndex int        `json:"callIndex"`
}

type PoolEventFlow struct {
	BaseFlow  float64 `json:"baseFlow"`
	QuoteFlow float64 `json:"quoteFlow"`
}

type PoolEventDescriptor struct {
	EntityType   tables.EntityType `json:"entityType"`
	ChangeType   tables.ChangeType `json:"changeType"`
	PositionType tables.PosType    `json:"positionType"`
}

type PoolRangeFields struct {
	BidTick   int  `json:"bidTick"`
	AskTick   int  `json:"askTick"`
	IsBuy     bool `json:"isBuy"`
	InBaseQty bool `json:"inBaseQty"`
}

func (t PoolTxEvent) Time() int {
	return t.TxTime
}

func (p PoolTxEvent) Hash(buf *bytes.Buffer) [32]byte {
	if buf == nil {
		buf = new(bytes.Buffer)
	} else {
		buf.Reset()
	}
	buf.Grow(270)
	binary.Write(buf, binary.BigEndian, int32(p.BlockNum))
	buf.WriteString(string(p.TxHash))
	binary.Write(buf, binary.BigEndian, int32(p.TxTime))
	buf.WriteString(string(p.User))

	buf.WriteString(string(p.ChainId))
	buf.WriteString(string(p.Base))
	buf.WriteString(string(p.Quote))
	binary.Write(buf, binary.BigEndian, int32(p.PoolIdx))

	binary.Write(buf, binary.BigEndian, int32(p.EntityType))
	binary.Write(buf, binary.BigEndian, int32(p.ChangeType))
	binary.Write(buf, binary.BigEndian, int32(p.PositionType))

	binary.Write(buf, binary.BigEndian, int32(p.BidTick))
	binary.Write(buf, binary.BigEndian, int32(p.AskTick))
	binary.Write(buf, binary.BigEndian, p.IsBuy)
	binary.Write(buf, binary.BigEndian, p.InBaseQty)
	return sha256.Sum256(buf.Bytes())
}
