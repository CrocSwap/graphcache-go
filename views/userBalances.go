package views

import (
	"github.com/CrocSwap/graphcache-go/types"
)

type UserBalanceResponse struct {
	ChainId types.ChainId       `json:"chainId"`
	User    types.EthAddress    `json:"user"`
	Block   int64               `json:"block"`
	Tokens  []types.UserBalance `json:"tokens"`
}

func (v *Views) QueryUserBalances(chainId types.ChainId, user types.EthAddress) (UserBalanceResponse, error) {
	resp := UserBalanceResponse{
		ChainId: chainId,
		User:    user,
		Block:   v.Models.LatestBlock(chainId),
		Tokens:  make([]types.UserBalance, 0),
	}

	balances := v.Models.RetrieveUserBalances(chainId, user)
	for _, bal := range balances {
		resp.Tokens = append(resp.Tokens, bal)
	}

	return resp, nil
}
