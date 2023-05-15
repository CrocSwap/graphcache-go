package views

import (
	"github.com/CrocSwap/graphcache-go/types"
)

type UserBalanceResponse struct {
	chainId types.ChainId       `json:"chainId"`
	user    types.EthAddress    `json:"user"`
	block   int64               `json:"block"`
	tokens  []types.UserBalance `json:"tokens"`
}
