package tables

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
)

type UniSwapsTable struct{}

func (tbl UniSwapsTable) GetID(r AggEvent) string {
	return r.ID
}

func (tbl UniSwapsTable) GetTime(r AggEvent) int {
	return r.Time
}



type UniSwapSubGraph struct {
	ID              string `json:"id"`
	Transaction	 struct {
		ID string `json:"id"`
		BlockNumber string `json:"blockNumber"`
	}`json:"transaction"`

	// EventIndex      int    `json:"eventIndex"`
	Pool            struct {
		ID string `json:"id"`
		Token0 struct {
			ID string `json:"id"`
			Symbol string `json:"symbol"`
		} `json:"token0"`
		Token1 struct {
			ID string `json:"id"`
			Symbol string `json:"symbol"`
		} `json:"token1"`
		
	} `json:"pool"`
	Sender string `json:"sender"`
	Recipient string `json:"recipient"`
	Amount0      string `json:"amount0"`
	Amount1     string `json:"amount1"`
	Timestamp string `json:"timestamp"`

}



type UniSwapSubGraphData struct {
	AggEvents []UniSwapSubGraph `json:"swaps"`
}

type UniSwapSubGraphResp struct {
	Data UniSwapSubGraphData `json:"data"`
}

func (tbl UniSwapsTable) ConvertSubGraphRow(r UniSwapSubGraph, network string) AggEvent {
	amount0 := parseNullableFloat64(r.Amount0)
	amount1 := parseNullableFloat64(r.Amount1)

	base := r.Pool.Token0.ID
	quote := r.Pool.Token1.ID
	baseFlow:= *amount0
	quoteFlow := *amount1


	if strings.ToLower(base) < strings.ToLower(base) {
		base, quote = quote, base
		baseFlow, quoteFlow = quoteFlow, baseFlow
	}

	price := math.Abs(baseFlow / quoteFlow)
	// convert price to string
	priceString := fmt.Sprintf("%f", price)

	// convert pool.Id to int


	return AggEvent{
		ID:            network + r.ID,
		EventIndex:    0,
		Network:       network,
		TX:            r.Transaction.ID,
		Base:          base,
		Quote:         quote,
		PoolIdx:       36000,
		PoolHash:      r.Pool.ID + base + quote,
		Block:         parseInt(r.Transaction.BlockNumber),
		Time:          parseInt(r.Timestamp),
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
		BaseFlow:      baseFlow,
		QuoteFlow:     quoteFlow,
		Source:        "graph",
	}
}

func (tbl UniSwapsTable) ParseSubGraphResp(body []byte) ([]UniSwapSubGraph, error) {
	var parsed UniSwapSubGraphResp

	err := json.Unmarshal(body, &parsed)
	if err != nil {
		return nil, err
	}

	ret := make([]UniSwapSubGraph, 0)
	for _, entry := range parsed.Data.AggEvents {
		ret = append(ret, entry)
	}
	return ret, nil
}

func (tbl UniSwapsTable) SqlTableName() string { return "aggevents" }
