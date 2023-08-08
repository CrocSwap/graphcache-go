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

To run the repo in such a way that only the swaps from uniswap syncs and no other data is pulled, perform the following steps. This is meant to be run alongside the normal implementation of graphcache-go to supplement candles from other pools and historical data. Candles will be found at `gcgo/pool_candles` when run in this mode.

1. Create a .env file and add the vars:

```
UNISWAP_CANDLES=true // Flag to put system into Uniswap Candles mode
UNISWAP_DAYS_OF_CANDLES_BEFORE_SERVER_READY=30 //
UNISWAP_HOUR_TO_SYNC_SHARDS=1
UNISWAP_GCS_BUCKET_NAME=gcgo-swap-shards
UNISWAP_SHARDS_PATH=./db/shards
UNISWAP_PATH_TO_GCS_CREDENTIALS=./GCS_credentials.json
```

2. Add the credentials file `GCS_credentials.json`
3. Build and run the container: `docker-compose -f ./docker-compose.uniswap.yml up`
4. To run w/o docker - `go build && ./graphcache-go` will work assuming you have the right env.

On startup of the server, a few things will happen

1. the Polling Syncer will begin pulling data from the uniswap subgraph into memory every minute or so.
2. It will then iterate all the dance from today back to January first and attempt to load them either from a GCS Shard or from the subgraph.
3. A task `SyncLocalShardsWithUniswap` is run once per day to create any shards and store them in GCS for a faster reboot later.

#### Env Explanation

UNISWAP_CANDLES: Flag to put system into Uniswap Candles mode
UNISWAP_DAYS_OF_CANDLES_BEFORE_SERVER_READY: Don't expose endpoints until this many days have been ingested into memory
UNISWAP_HOUR_TO_SYNC_SHARDS=Hour to run sync task, 1 => 1AM, 13 => 1PM
