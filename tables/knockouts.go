package tables

import (
	"database/sql"
	"encoding/json"
	"log"
	"strings"
)

type KnockoutTable struct{}

func (tbl KnockoutTable) GetID(r KnockoutCross) string {
	return r.ID
}

func (tbl KnockoutTable) GetTime(r KnockoutCross) int {
	return r.Time
}

func (tbl KnockoutTable) GetBlock(r KnockoutCross) int {
	return r.Block
}

type KnockoutCross struct {
	ID         string  `json:"id" db:"id"`
	Network    string  `json:"network" db:"network"`
	Tx         string  `json:"tx" db:"tx"`
	Block      int     `json:"block" db:"block"`
	Time       int     `json:"time" db:"time"`
	Base       string  `json:"base" db:"base"`
	Quote      string  `json:"quote" db:"quote"`
	PoolIdx    int     `json:"poolIdx" db:"poolIdx"`
	PoolHash   string  `json:"poolHash" db:"poolHash"`
	Tick       int     `json:"tick" db:"tick"`
	IsBid      int     `json:"isBid" db:"isBid"`
	PivotTime  int     `json:"pivotTime" db:"pivotTime"`
	FeeMileage float64 `json:"feeMileage" db:"feeMileage"`
}

type KnockoutCrossSubGraph struct {
	ID              string `json:"id"`
	TransactionHash string `json:"transactionHash"`
	Pool            struct {
		ID      string `json:"id"`
		Base    string `json:"base"`
		Quote   string `json:"quote"`
		PoolIdx string `json:"poolIdx"`
	} `json:"pool"`
	Block      string `json:"block"`
	Time       string `json:"time"`
	Tick       int    `json:"tick"`
	IsBid      bool   `json:"isBid"`
	PivotTime  string `json:"pivotTime"`
	FeeMileage string `json:"feeMileage"`
}

type KnockoutCrossSubGraphData struct {
	KnockoutCrosses []KnockoutCrossSubGraph `json:"knockoutCrosses"`
}

type KnockoutCrossSubGraphResp struct {
	Data KnockoutCrossSubGraphData `json:"data"`
}

func (tbl KnockoutTable) ConvertSubGraphRow(r KnockoutCrossSubGraph, network string) KnockoutCross {
	base, quote := r.Pool.Base, r.Pool.Quote

	// Flip is base/quote is actually reversed
	if strings.ToLower(r.Pool.Base) > strings.ToLower(r.Pool.Quote) {
		base, quote = quote, base
	}

	return KnockoutCross{
		ID:         r.ID + network,
		Network:    network,
		Tx:         r.TransactionHash,
		Block:      parseInt(r.Block),
		Time:       parseInt(r.Time),
		Base:       base,
		Quote:      quote,
		PoolIdx:    parseInt(r.Pool.PoolIdx),
		PoolHash:   hashPool(base, quote, parseInt(r.Pool.PoolIdx)),
		Tick:       r.Tick,
		IsBid:      boolToInt(r.IsBid),
		PivotTime:  parseInt(r.PivotTime),
		FeeMileage: *parseNullableFloat64(r.FeeMileage),
	}
}

func (tbl KnockoutTable) SqlTableName() string { return "knockout_crosses" }

func (tbl KnockoutTable) ReadSqlRow(rows *sql.Rows) KnockoutCross {
	var cross KnockoutCross
	err := rows.Scan(
		&cross.ID,
		&cross.Network,
		&cross.Tx,
		&cross.Block,
		&cross.Time,
		&cross.Base,
		&cross.Quote,
		&cross.PoolIdx,
		&cross.PoolHash,
		&cross.Tick,
		&cross.IsBid,
		&cross.PivotTime,
		&cross.FeeMileage,
	)
	if err != nil {
		log.Fatal(err.Error())
	}
	return cross
}

func (tbl KnockoutTable) ParseSubGraphResp(body []byte) ([]KnockoutCrossSubGraph, error) {
	var parsed KnockoutCrossSubGraphResp

	err := json.Unmarshal(body, &parsed)
	if err != nil {
		return nil, err
	}

	ret := make([]KnockoutCrossSubGraph, 0)
	for _, entry := range parsed.Data.KnockoutCrosses {
		ret = append(ret, entry)
	}
	return ret, nil
}
