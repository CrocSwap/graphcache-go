# graphcache-go

Provides simple and fast endpoints to index the CrocSwap protocol on Ethereum based networks.

## Quickstart

To compile, from the project root directory call

`go build`

After building to run, from the project root directory call

`./graphcache-go`

## Network options

By default the instance uses the network config at `./config/networks.json`. To use a different configuration file run with

`./graphcache-go -netCfg [NETWORK_CONFIG_PATH]`

The RPC endpoint can be overriden in the environment by setting the `RPC_MAINNET` env variable before running:

    export RPC_MAINNET=[RPC_URL]
    ./graphcache-go

## Endpoints

The following exposed endpoints and their URL and paramters are listed in `server/server.go`

- `gcgo/user_balance_tokens` - List all tokens the user has potential surplus collateral
- `gcgo/user_positions` - List all concentrated and ambient liquidity positions
- `gcgo/pool_positions` - List N most recent concentrated and ambient positions in a pool
- `gcgo/pool_position_apy_leaders` - List top N positions in pool by annualized fee APY
- `gcgo/user_pool_positions` - List liquidity positions of a user in a single pool
- `gcgo/position_stats` - Describe a single liquidity position
- `gcgo/user_limit_orders` - List all non-zero knockout liquidity positions of a user
- `gcgo/pool_limit_orders` - List N most recent knockout liquidity position in a pool
- `gcgo/user_pool_limit_orders` - List knockout positions of a user in a single pool
- `gcgo/limit_stats` - Describe a single knockout position
- `gcgo/user_txs` - List all dex trading transactions of a user
- `gcgo/pool_txs` - List N most recent trading transactions in a pool
- `gcgo/pool_liq_curve` - Return the most recent description of the liquidity curve in a pool

## Uniswap Candles

Create a .env file and add the var: `UNISWAP_CANDLES=true`
This will run the repo in such a way that only the swaps from uniswap syncs and no other data is pulled. This is meant to be run alongside the normal implementation of graphcache-go to supplement candles from other pools and historical data. Candles will be found at the same endpoint when run in this mode.

On startup of the server, a few things will happen

1. the Polling Syncer will begin pulling data from the uniswap subgraph into memory every minute or so.
2. Concurrently, it will fetch the date of the most recent swap in the db (`dbLast`) and then try to update the db to the start time from the uniswap subgraph while producing candle data for those swaps. It goes in forward order.
3. Once the db is caught up to start time, it will go in reverse order from `dbLast` until Jan1 loading candle data from the db.

## Run Uniswap Candles

1. Pull db w/ swaps from January 1, 2023 `./_data/pull-db.sh`
2. Build and run the container: `docker-compose up`
