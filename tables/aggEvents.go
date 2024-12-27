package tables

import (
	"strings"

	stdjson "encoding/json"

	"github.com/goccy/go-json"
)

type AggEventsTable struct{}

func (tbl AggEventsTable) GetID(r AggEvent) string {
	return r.ID
}

func (tbl AggEventsTable) GetTime(r AggEvent) int {
	return r.Time
}

func (tbl AggEventsTable) GetBlock(r AggEvent) int {
	return r.Block
}

type AggEvent struct {
	ID            string  `json:"id" db:"id"`
	Network       string  `json:"network" db:"network"`
	Base          string  `json:"base" db:"base"`
	Quote         string  `json:"quote" db:"quote"`
	PoolIdx       int     `json:"poolIdx" db:"poolIdx"`
	Block         int     `json:"block" db:"block"`
	Time          int     `json:"time" db:"time"`
	BidTick       int     `json:"bidTick" db:"bidTick"`
	AskTick       int     `json:"askTick" db:"askTick"`
	IsFeeChange   bool    `json:"isFeeChange" db:"isFeeChange"`
	IsSwap        bool    `json:"isSwap" db:"isSwap"`
	IsLiq         bool    `json:"isLiq" db:"isLiq"`
	IsTickSkewed  bool    `json:"isTickSkewed" db:"isTickSkewed"`
	FlowsAtMarket bool    `json:"flowsAtMarket" db:"flowsAtMarket"`
	InBaseQty     bool    `json:"inBaseQty" db:"inBaseQty"`
	BaseFlow      float64 `json:"baseFlow" db:"baseFlow"`
	QuoteFlow     float64 `json:"quoteFlow" db:"quoteFlow"`
	FeeRate       int     `json:"feeRate" db:"feeRate"`
	EventIndex    int     `json:"eventIndex" db:"eventIndex"`
}

type AggEventSubGraph struct {
	ID            string       `json:"id"`
	Pool          SubGraphPool `json:"pool"`
	Block         string       `json:"block"`
	Time          string       `json:"time"`
	BidTick       int          `json:"bidTick"`
	AskTick       int          `json:"askTick"`
	InBaseQty     bool         `json:"inBaseQty"`
	IsSwap        bool         `json:"isSwap"`
	IsLiq         bool         `json:"isLiq"`
	IsFeeChange   bool         `json:"isFeeChange"`
	IsTickSkewed  bool         `json:"isTickSkewed"`
	FlowsAtMarket bool         `json:"flowsAtMarket"`
	BaseFlow      string       `json:"baseFlow"`
	QuoteFlow     string       `json:"quoteFlow"`
	FeeRate       int          `json:"feeRate"`
	EventIndex    int          `json:"eventIndex"`
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

	// strings.Clone is needed because go-json doesn't let go of the original
	// buffers, which leads to gigabytes of wasted RAM compared to std json.
	return AggEvent{
		ID:            network + r.ID,
		Network:       network,
		Base:          strings.Clone(base),
		Quote:         strings.Clone(quote),
		PoolIdx:       parseInt(r.Pool.PoolIdx),
		Block:         parseInt(r.Block),
		Time:          parseInt(r.Time),
		BidTick:       r.BidTick,
		AskTick:       r.AskTick,
		IsFeeChange:   r.IsFeeChange,
		IsLiq:         r.IsLiq,
		IsSwap:        r.IsSwap,
		IsTickSkewed:  r.IsTickSkewed,
		InBaseQty:     r.InBaseQty,
		FlowsAtMarket: r.FlowsAtMarket,
		FeeRate:       r.FeeRate,
		BaseFlow:      *baseFlow,
		QuoteFlow:     *quoteFlow,
		EventIndex:    r.EventIndex,
	}
}

func (tbl AggEventsTable) ParseSubGraphResp(body []byte) ([]AggEventSubGraph, error) {
	var parsed AggEventSubGraphResp

	err := stdjson.Unmarshal(body, &parsed)
	if err != nil {
		return nil, err
	}

	ret := make([]AggEventSubGraph, 0, len(parsed.Data.AggEvents))
	for _, entry := range parsed.Data.AggEvents {
		ret = append(ret, entry)
	}
	return ret, nil
}

func (tbl AggEventsTable) ParseSubGraphRespUnwrapped(body []byte) ([]AggEventSubGraph, error) {
	var parsed AggEventSubGraphData

	err := json.Unmarshal(body, &parsed.AggEvents)
	if err != nil {
		return nil, err
	}

	ret := make([]AggEventSubGraph, 0, len(parsed.AggEvents))
	for _, entry := range parsed.AggEvents {
		ret = append(ret, entry)
	}
	return ret, nil
}

func (tbl AggEventsTable) SqlTableName() string { return "aggevents" }
