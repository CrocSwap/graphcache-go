package views

import (
	"github.com/CrocSwap/graphcache-go/types"
)

type UserTokensResponse struct {
	ChainId types.ChainId      `json:"chainId"`
	User    types.EthAddress   `json:"user"`
	Block   int64              `json:"block"`
	Tokens  []types.EthAddress `json:"tokens"`
}

func (v *Views) QueryUserTokens(chainId types.ChainId, user types.EthAddress) (UserTokensResponse, error) {
	resp := UserTokensResponse{
		ChainId: chainId,
		User:    user,
		Block:   v.Cache.LatestBlock(chainId),
		Tokens:  make([]types.EthAddress, 0),
	}

	balances := v.Cache.RetrieveUserBalances(chainId, user)

	for _, bal := range balances {
		resp.Tokens = append(resp.Tokens, bal)
	}

	return resp, nil
}
