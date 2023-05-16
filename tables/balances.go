package tables

import (
	"database/sql"
	"encoding/json"
	"log"
	"strconv"
)

type BalanceTable struct {
}

func (tbl BalanceTable) GetID(r Balance) string {
	return r.ID
}

func (tbl BalanceTable) GetTime(r Balance) int {
	return r.Time
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
		Tx:      r.TransactionHash,
		Block:   stringNum(r.Block),
		Time:    stringNum(r.Time),
		User:    r.User,
		Token:   r.Token,
	}
}

func stringNum(val string) int {
	ret, err := strconv.Atoi(val)
	if err != nil {
		log.Fatal("Subgraph number conversion error")
	}
	return ret
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

	err := json.Unmarshal(body, &parsed)
	if err != nil {
		return nil, err
	}

	ret := make([]BalanceSubGraph, 0)
	for _, entry := range parsed.Data.UserBalances {
		ret = append(ret, entry)
	}
	return ret, nil
}
