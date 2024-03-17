package tables

import (
	"database/sql"
	"encoding/json"
	"log"
	"strings"
)

type SwapsTable struct{}

func (tbl SwapsTable) GetID(r Swap) string {
	return r.ID
}

func (tbl SwapsTable) GetTime(r Swap) int {
	return r.Time
}

func (tbl SwapsTable) GetBlock(r Swap) int {
	return r.Block
}

type Swap struct {
	ID         string   `json:"id" db:"id"`
	CallIndex  int      `json:"callIndex" db:"callIndex"`
	Network    string   `json:"network" db:"network"`
	TX         string   `json:"tx" db:"tx"`
	User       string   `json:"user" db:"user"`
	Block      int      `json:"block" db:"block"`
	Time       int      `json:"time" db:"time"`
	Base       string   `json:"base" db:"base"`
	Quote      string   `json:"quote" db:"quote"`
	PoolIdx    int      `json:"poolIdx" db:"poolIdx"`
	PoolHash   string   `json:"poolHash" db:"poolHash"`
	IsBuy      int      `json:"isBuy" db:"isBuy"`
	InBaseQty  int      `json:"inBaseQty" db:"inBaseQty"`
	Qty        float64  `json:"qty" db:"qty"`
	LimitPrice *float64 `json:"limitPrice" db:"limitPrice"`
	MinOut     *float64 `json:"minOut" db:"minOut"`
	BaseFlow   float64  `json:"baseFlow" db:"baseFlow"`
	QuoteFlow  float64  `json:"quoteFlow" db:"quoteFlow"`
}

type SwapSubGraph struct {
	ID              string `json:"id"`
	TransactionHash string `json:"transactionHash"`
	CallIndex       int    `json:"callIndex"`
	User            string `json:"user"`
	Pool            struct {
		Base    string `json:"base"`
		Quote   string `json:"quote"`
		PoolIdx string `json:"poolIdx"`
	} `json:"pool"`
	Block      string `json:"block"`
	Time       string `json:"time"`
	IsBuy      bool   `json:"isBuy"`
	InBaseQty  bool   `json:"inBaseQty"`
	Qty        string `json:"qty"`
	LimitPrice string `json:"limitPrice"`
	MinOut     string `json:"minOut"`
	BaseFlow   string `json:"baseFlow"`
	QuoteFlow  string `json:"quoteFlow"`
}

type SwapSubGraphData struct {
	Swaps []SwapSubGraph `json:"swaps"`
}

type SwapSubGraphResp struct {
	Data SwapSubGraphData `json:"data"`
}

func (tbl SwapsTable) ConvertSubGraphRow(r SwapSubGraph, network string) Swap {
	base, quote := r.Pool.Base, r.Pool.Quote
	baseFlow := parseNullableFloat64(r.BaseFlow)
	quoteFlow := parseNullableFloat64(r.QuoteFlow)

	// Flip is base/quote is actually reversed
	if strings.ToLower(r.Pool.Base) > strings.ToLower(r.Pool.Quote) {
		base, quote = quote, base
		baseFlow, quoteFlow = quoteFlow, baseFlow
	}

	return Swap{
		ID:         network + r.ID,
		CallIndex:  r.CallIndex,
		Network:    network,
		TX:         r.TransactionHash,
		User:       r.User,
		Block:      parseInt(r.Block),
		Time:       parseInt(r.Time),
		Base:       base,
		Quote:      quote,
		PoolIdx:    parseInt(r.Pool.PoolIdx),
		PoolHash:   hashPool(base, quote, parseInt(r.Pool.PoolIdx)),
		IsBuy:      boolToInt(r.IsBuy),
		InBaseQty:  boolToInt(r.InBaseQty),
		Qty:        *parseNullableFloat64(r.Qty),
		LimitPrice: parseNullableFloat64(r.LimitPrice),
		MinOut:     parseNullableFloat64(r.MinOut),
		BaseFlow:   *baseFlow,
		QuoteFlow:  *quoteFlow,
	}
}

func (tbl SwapsTable) SqlTableName() string { return "swaps" }

func (tbl SwapsTable) ReadSqlRow(rows *sql.Rows) Swap {
	var swap Swap
	err := rows.Scan(
		&swap.ID,
		&swap.CallIndex,
		&swap.Network,
		&swap.TX,
		&swap.User,
		&swap.Block,
		&swap.Time,
		&swap.Base,
		&swap.Quote,
		&swap.PoolIdx,
		&swap.PoolHash,
		&swap.IsBuy,
		&swap.InBaseQty,
		&swap.Qty,
		&swap.LimitPrice,
		&swap.MinOut,
		&swap.BaseFlow,
		&swap.QuoteFlow,
	)
	if err != nil {
		log.Fatal(err.Error())
	}
	return swap
}

func (tbl SwapsTable) ParseSubGraphResp(body []byte) ([]SwapSubGraph, error) {
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
