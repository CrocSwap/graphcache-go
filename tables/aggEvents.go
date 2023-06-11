package tables

import (
	"encoding/json"
	"strings"
)

type AggEventsTable struct{}

func (tbl AggEventsTable) GetID(r AggEvent) string {
	return r.ID
}

func (tbl AggEventsTable) GetTime(r AggEvent) int {
	return r.Time
}

type AggEvent struct {
	ID           string  `json:"id" db:"id"`
	EventIndex   int     `json:"eventIndex" db:"eventIndex"`
	Network      string  `json:"network" db:"network"`
	TX           string  `json:"tx" db:"tx"`
	Base         string  `json:"base" db:"base"`
	Quote        string  `json:"quote" db:"quote"`
	PoolIdx      int     `json:"poolIdx" db:"poolIdx"`
	PoolHash     string  `json:"poolHash" db:"poolHash"`
	User         string  `json:"user" db:"user"`
	Block        int     `json:"block" db:"block"`
	Time         int     `json:"time" db:"time"`
	BidTick      int     `json:"bidTick" db:"bidTick"`
	AskTick      int     `json:"askTick" db:"askTick"`
	SwapPrice    string  `json:"swapPrice"`
	IsFeeChange  bool    `json:"isFeeChange" db:"isFeeChange"`
	IsSwap       bool    `json:"isSwap" db:"isSwap"`
	IsLiq        bool    `json:"isLiq" db:"isLiq"`
	IsTickSkewed bool    `json:"isTickSkewed" db:"isTickSkewed"`
	InBaseQty    bool    `json:"inBaseQty" db:"inBaseQty"`
	BaseFlow     float64 `json:"baseFlow" db:"baseFlow"`
	QuoteFlow    float64 `json:"quoteFlow" db:"quoteFlow"`
	FeeRate      int     `json:"feeRate" db:"feeRate"`
	Source       string  `json:"source" db:"source"`
}

type AggEventSubGraph struct {
	ID              string `json:"id"`
	TransactionHash string `json:"transactionHash"`
	EventIndex      int    `json:"eventIndex"`
	Pool            struct {
		ID      string `json:"id"`
		Base    string `json:"base"`
		Quote   string `json:"quote"`
		PoolIdx string `json:"poolIdx"`
	} `json:"pool"`
	Block        string `json:"block"`
	Time         string `json:"time"`
	BidTick      int    `json:"bidTick"`
	AskTick      int    `json:"askTick"`
	SwapPrice    string `json:"swapPrice"`
	InBaseQty    bool   `json:"inBaseQty"`
	IsSwap       bool   `json:"isSwap"`
	IsLiq        bool   `json:"isLiq"`
	IsFeeChange  bool   `json:"isFeeChange"`
	IsTickSkewed bool   `json:"isTickSkewed" db:"isTickSkewed"`
	BaseFlow     string `json:"baseFlow"`
	QuoteFlow    string `json:"quoteFlow"`
	FeeRate      int    `json:"feeRate"`
}

type AggEventSubGraphData struct {
	AggEvents []AggEventSubGraph `json:"aggEvents"`
}

type AggEventSubGraphResp struct {
	Data AggEventSubGraphData `json:"data"`
}

func (tbl AggEventsTable) ConvertSubGraphRow(r AggEventSubGraph, network string) AggEvent {
	base, quote := r.Pool.Base, r.Pool.Quote
	baseFlow := parseNullableFloat64(r.BaseFlow)
	quoteFlow := parseNullableFloat64(r.QuoteFlow)

	// Flip is base/quote is actually reversed
	if strings.ToLower(r.Pool.Base) > strings.ToLower(r.Pool.Quote) {
		base, quote = quote, base
		baseFlow, quoteFlow = quoteFlow, baseFlow
	}

	return AggEvent{
		ID:          network + r.ID,
		EventIndex:  r.EventIndex,
		Network:     network,
		TX:          r.TransactionHash,
		Base:        base,
		Quote:       quote,
		PoolIdx:     parseInt(r.Pool.PoolIdx),
		PoolHash:    hashPool(base, quote, parseInt(r.Pool.PoolIdx)),
		Block:       parseInt(r.Block),
		Time:        parseInt(r.Time),
		BidTick:     r.BidTick,
		AskTick:     r.AskTick,
		SwapPrice:   r.SwapPrice,
		IsFeeChange: r.IsFeeChange,
		IsLiq:       r.IsLiq,
		IsSwap:      r.IsSwap,
		FeeRate:     r.FeeRate,
		BaseFlow:    parseFloat(r.BaseFlow),
		QuoteFlow:   parseFloat(r.QuoteFlow),
		Source:      "graph",
	}
}

func (tbl AggEventsTable) ParseSubGraphResp(body []byte) ([]AggEventSubGraph, error) {
	var parsed AggEventSubGraphResp

	err := json.Unmarshal(body, &parsed)
	if err != nil {
		return nil, err
	}

	ret := make([]AggEventSubGraph, 0)
	for _, entry := range parsed.Data.AggEvents {
		ret = append(ret, entry)
	}
	return ret, nil
}

func (tbl AggEventsTable) SqlTableName() string { return "aggevents" }
