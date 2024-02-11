package types

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
)

type PoolTxEvent struct {
	EthTxHeader
	PoolLocation
	PoolEventFlow
	PoolEventDescriptor
	PoolRangeFields
}

type EthTxHeader struct {
	BlockNum int        `json:"blockNum"`
	TxHash   EthTxHash  `json:"txHash"`
	TxTime   int        `json:"txTime"`
	User     EthAddress `json:"user"`
}

type PoolEventFlow struct {
	BaseFlow  float64 `json:"baseFlow"`
	QuoteFlow float64 `json:"quoteFlow"`
}

type PoolEventDescriptor struct {
	EntityType   string `json:"entityType"`
	ChangeType   string `json:"changeType"`
	PositionType string `json:"positionType"`
}

type PoolRangeFields struct {
	BidTick   int  `json:"bidTick"`
	AskTick   int  `json:"askTick"`
	IsBuy     bool `json:"isBuy"`
	InBaseQty bool `json:"inBaseQty"`
}

func (p PoolTxEvent) Hash() [32]byte {
	buf := new(bytes.Buffer)
	buf.Grow(270)
	binary.Write(buf, binary.BigEndian, int32(p.BlockNum))
	buf.WriteString(string(p.TxHash))
	binary.Write(buf, binary.BigEndian, int32(p.TxTime))
	buf.WriteString(string(p.User))

	buf.WriteString(string(p.ChainId))
	buf.WriteString(string(p.Base))
	buf.WriteString(string(p.Quote))
	binary.Write(buf, binary.BigEndian, int32(p.PoolIdx))

	buf.WriteString(string(p.EntityType))
	buf.WriteString(string(p.ChangeType))
	buf.WriteString(string(p.PositionType))

	binary.Write(buf, binary.BigEndian, int32(p.BidTick))
	binary.Write(buf, binary.BigEndian, int32(p.AskTick))
	binary.Write(buf, binary.BigEndian, p.IsBuy)
	binary.Write(buf, binary.BigEndian, p.InBaseQty)
	return sha256.Sum256(buf.Bytes())
}
