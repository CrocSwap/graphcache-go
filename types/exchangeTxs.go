package types

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
