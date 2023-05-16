package types

import "math/big"

type UserBalance struct {
	Token   EthAddress `json:"token"`
	Balance big.Int    `json:"balance"`
}

type LiqPosition struct {
	ChainID      string         `json:"chainId"`
	TX           EthTxHash      `json:"tx"`
	Base         EthAddress     `json:"base"`
	Quote        EthAddress     `json:"quote"`
	PoolIdx      int            `json:"poolId"`
	PoolHash     EthTxHash      `json:"poolHash"`
	User         EthAddress     `json:"user"`
	Block        int            `json:"block"`
	Time         int            `json:"time"`
	PositionType string         `json:"positionType"`
	BidTick      int            `json:"bidTick"`
	AskTick      int            `json:"askTick"`
	IsBid        bool           `json:"isBid"`
	PosSlot      EthStorageHash `json:"positionStorageSlot"`
}
