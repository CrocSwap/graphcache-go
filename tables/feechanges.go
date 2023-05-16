package tables

import (
	"database/sql"
	"encoding/json"
	"log"
	"strings"
)

type FeeTable struct{}

func (tbl FeeTable) GetID(r FeeChange) string {
	return r.ID
}

func (tbl FeeTable) GetTime(r FeeChange) int {
	return r.Time
}

type FeeChange struct {
	ID        string `json:"id" db:"id"`
	CallIndex int    `json:"callIndex" db:"callIndex"`
	Network   string `json:"network" db:"network"`
	Tx        string `json:"tx" db:"tx"`
	Block     int    `json:"block" db:"block"`
	Time      int    `json:"time" db:"time"`
	Base      string `json:"base" db:"base"`
	Quote     string `json:"quote" db:"quote"`
	PoolIdx   int    `json:"poolIdx" db:"poolIdx"`
	PoolHash  string `json:"poolHash" db:"poolHash"`
	FeeRate   int    `json:"feeRate" db:"feeRate"`
}

type FeeChangeSubGraph struct {
	ID              string `json:"id"`
	TransactionHash string `json:"transactionHash"`
	CallIndex       int    `json:"callIndex"`
	Block           string `json:"block"`
	Time            string `json:"time"`
	Pool            struct {
		ID      string `json:"id"`
		Base    string `json:"base"`
		Quote   string `json:"quote"`
		PoolIdx string `json:"poolIdx"`
	} `json:"pool"`
	FeeRate int `json:"feeRate"`
}

type FeeChangeSubGraphData struct {
	FeeChanges []FeeChangeSubGraph `json:"feeChanges"`
}

type FeeChangeSubGraphResp struct {
	Data FeeChangeSubGraphData `json:"data"`
}

func (tbl FeeTable) ConvertSubGraphRow(r FeeChangeSubGraph, network string) FeeChange {
	base, quote := r.Pool.Base, r.Pool.Quote

	// Flip is base/quote is actually reversed
	if strings.ToLower(r.Pool.Base) > strings.ToLower(r.Pool.Quote) {
		base, quote = quote, base
	}

	return FeeChange{
		ID:        r.ID + network,
		CallIndex: r.CallIndex,
		Network:   network,
		Tx:        r.TransactionHash,
		Block:     parseInt(r.Block),
		Time:      parseInt(r.Time),
		Base:      base,
		Quote:     quote,
		PoolIdx:   parseInt(r.Pool.PoolIdx),
		PoolHash:  hashPool(base, quote, parseInt(r.Pool.PoolIdx)),
		FeeRate:   r.FeeRate,
	}
}

func (tbl FeeTable) SqlTableName() string { return "fee_changes" }

func (tbl FeeTable) ReadSqlRow(rows *sql.Rows) FeeChange {
	var feeChange FeeChange
	err := rows.Scan(
		&feeChange.ID,
		&feeChange.CallIndex,
		&feeChange.Network,
		&feeChange.Tx,
		&feeChange.Block,
		&feeChange.Time,
		&feeChange.Base,
		&feeChange.Quote,
		&feeChange.PoolIdx,
		&feeChange.PoolHash,
		&feeChange.FeeRate,
	)
	if err != nil {
		log.Fatal(err.Error())
	}
	return feeChange
}

func (tbl FeeTable) ParseSubGraphResp(body []byte) ([]FeeChangeSubGraph, error) {
	var parsed FeeChangeSubGraphResp

	err := json.Unmarshal(body, &parsed)
	if err != nil {
		return nil, err
	}

	ret := make([]FeeChangeSubGraph, 0)
	for _, entry := range parsed.Data.FeeChanges {
		ret = append(ret, entry)
	}
	return ret, nil
}
