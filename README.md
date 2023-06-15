# graphcache-go

To compile, from the project root directory call

`go build`

After building to run, from the project root directory call

`./graphcache-go`

## Network options

By default the instance uses the network config at `./config/networks.json`. To use a different configuration file run with

`./graphcache-go -netCfg [NETWORK_CONFIG_PATH]`

The RPC endpoint can be overriden in the environment by setting the `RPC_MAINNET` env variable before running:

`export RPC_MAINNET=[RPC_URL]; ./graphcache-go`

## Endpoints

The following exposed endpoints and their URL and paramters are listed in `server/server.go`

* `gcgo/user_balance_tokens` - List all tokens the user has potential surplus collateral
* `gcgo/user_positions` - List all concentrated and ambient liquidity positions
* `gcgo/pool_positions` - List N most recent concentrated and ambient positions in a pool
* `gcgo/pool_position_apy_leaders` - List top N positions in pool by annualized fee APY
* `gcgo/user_pool_positions` - List liquidity positions of a user in a single pool
* `gcgo/position_stats` - Describe a single liquidity position
* `gcgo/user_limit_orders` - List all non-zero knockout liquidity positions of a user
* `gcgo/pool_limit_orders` - List N most recent knockout liquidity position in a pool
* `gcgo/user_pool_limit_orders` - List knockout positions of a user in a single pool
* `gcgo/limit_stats` - Describe a single knockout position
* `gcgo/user_txs` - List all dex trading transactions of a user
* `gcgo/pool_txs` - List N most recent trading transactions in a pool
* `gcgo/pool_liq_curve` - Return the most recent description of the liquidity curve in a pool
