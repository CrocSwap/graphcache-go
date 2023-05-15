package tables

import (
	"database/sql"
	"log"
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

type BalanceFn func(bal Balance)

func LoadTokenBalancesSql(db *sql.DB, ingestFn func(Balance)) {
	// Query the "balances" table
	rows, err := db.Query("SELECT * FROM balances")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	// Iterate over the rows
	for rows.Next() {
		// Create a new Balance struct
		var balance Balance

		// Scan the row values into the Balance struct fields
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

	// Check for any errors during iteration
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
}
