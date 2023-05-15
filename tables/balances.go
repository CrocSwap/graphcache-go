package tables

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
)

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

func ConvertBalanceToSql(r BalanceSubGraph, network string) Balance {
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

func LoadTokenBalancesSql(db *sql.DB, network string, ingestFn func(Balance)) {
	query := fmt.Sprintf("SELECT * FROM balances WHERE network == '%s'", network)
	rows, err := db.Query(query)

	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
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

		ingestFn(balance)
	}

	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
}
