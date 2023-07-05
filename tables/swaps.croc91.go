package tables

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
)

type Croc91SwapsTable struct{}

func (tbl Croc91SwapsTable) GetID(r AggEvent) string {
	return r.ID
}

func (tbl Croc91SwapsTable) GetTime(r AggEvent) int {
	return r.Time
}

type Croc91SwapSubGraphData struct {
	Swaps []SwapSubGraph `json:"swaps"`
}

type Cros91SwapSubGraphResp struct {
	Data SwapSubGraphData `json:"data"`
}

func (tbl Croc91SwapsTable) ConvertSubGraphRow(r SwapSubGraph, network string) AggEvent {
	base, quote := r.Pool.Base, r.Pool.Quote
	baseFlow := parseNullableFloat64(r.BaseFlow)
	quoteFlow := parseNullableFloat64(r.QuoteFlow)

	// Flip is base/quote is actually reversed
	if strings.ToLower(r.Pool.Base) > strings.ToLower(r.Pool.Quote) {
		base, quote = quote, base
		baseFlow, quoteFlow = quoteFlow, baseFlow
	}
	price := math.Abs(*baseFlow / *quoteFlow)
	// convert price to string
	priceString := fmt.Sprintf("%f", price)

	return AggEvent{
		ID:            network + r.ID,
		EventIndex:    0,
		Network:       network,
		TX:            r.TransactionHash,
		Base:          base,
		Quote:         quote,
		PoolIdx:       36000,
		PoolHash:      r.Pool.ID + base + quote,
		Block:         parseInt(r.Block),
		Time:          parseInt(r.Time),
		BidTick:       0,
		AskTick:       0,
		SwapPrice:     priceString,
		IsFeeChange:   false,
		IsLiq:         false,
		IsSwap:        true,
		IsTickSkewed:  false,
		InBaseQty:     true,
		FlowsAtMarket: false,
		FeeRate:       0,
		BaseFlow:      *baseFlow,
		QuoteFlow:     *quoteFlow,
		Source:        "graph",
	}
}

func (tbl Croc91SwapsTable) ParseSubGraphResp(body []byte) ([]SwapSubGraph, error) {
	var parsed SwapSubGraphResp

	err := json.Unmarshal(body, &parsed)
	if err != nil {
		return nil, err
	}

	ret := make([]SwapSubGraph, 0)
	for _, entry := range parsed.Data.Swaps {
		ret = append(ret, entry)
	}
	return ret, nil
}

func (tbl Croc91SwapsTable) SqlTableName() string { return "aggevents" }
