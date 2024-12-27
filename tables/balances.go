package tables

import (
	"database/sql"
	"log"
	"strings"

	stdjson "encoding/json"

	"github.com/goccy/go-json"
)

type BalanceTable struct{}

func (tbl BalanceTable) GetID(r Balance) string {
	return r.ID
}

func (tbl BalanceTable) GetTime(r Balance) int {
	return r.Time
}

func (tbl BalanceTable) GetBlock(r Balance) int {
	return r.Block
}

type Balance struct {
	ID      string `db:"id"`
	Network string `db:"network"`
	Tx      string `db:"tx"`
	Block   int    `db:"block"`
	Time    int    `db:"time"`
	User    string `db:"user"`
	Token   string `db:"token"`
}

type BalanceSubGraph struct {
	ID              string `json:"id"`
	TransactionHash string `json:"transactionHash"`
	Block           string `json:"block"`
	Time            string `json:"time"`
	User            string `json:"user"`
	Token           string `json:"token"`
}

type BalanceSubGrapData struct {
	UserBalances []BalanceSubGraph `json:"userBalances"`
}

type BalanceSubGraphResp struct {
	Data BalanceSubGrapData `json:"data"`
}

func (tbl BalanceTable) ConvertSubGraphRow(r BalanceSubGraph, network string) Balance {
	return Balance{
		ID:      network + r.ID,
		Network: network,
		Tx:      strings.Clone(r.TransactionHash),
		Block:   parseInt(r.Block),
		Time:    parseInt(r.Time),
		User:    strings.Clone(translateUser(r.User, "")),
		Token:   strings.Clone(r.Token),
	}
}

func (tbl BalanceTable) SqlTableName() string { return "balances" }

func (tbl BalanceTable) ReadSqlRow(rows *sql.Rows) Balance {
	var balance Balance
	err := rows.Scan(
		&balance.ID,
		&balance.Network,
		&balance.Tx,
		&balance.Block,
		&balance.Time,
		&balance.User,
		&balance.Token,
	)
	if err != nil {
		log.Fatal(err)
	}
	return balance
}

func (tbl BalanceTable) ParseSubGraphResp(body []byte) ([]BalanceSubGraph, error) {
	var parsed BalanceSubGraphResp

	err := stdjson.Unmarshal(body, &parsed)
	if err != nil {
		return nil, err
	}

	ret := make([]BalanceSubGraph, 0, len(parsed.Data.UserBalances))
	for _, entry := range parsed.Data.UserBalances {
		ret = append(ret, entry)
	}
	return ret, nil
}

func (tbl BalanceTable) ParseSubGraphRespUnwrapped(body []byte) ([]BalanceSubGraph, error) {
	var parsed BalanceSubGrapData

	err := json.Unmarshal(body, &parsed.UserBalances)
	if err != nil {
		return nil, err
	}

	ret := make([]BalanceSubGraph, 0, len(parsed.UserBalances))
	for _, entry := range parsed.UserBalances {
		ret = append(ret, entry)
	}
	return ret, nil
}
