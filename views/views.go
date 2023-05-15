package views

import (
	"github.com/CrocSwap/graphcache-go/models"
	"github.com/CrocSwap/graphcache-go/types"
)

type IViews interface {
	QueryUserBalances(chainId types.ChainId, user types.EthAddress) (UserBalanceResponse, error)
}

type Views struct {
	Models *models.Models
}

func (v *Views) QueryUserBalances(chainId types.ChainId, user types.EthAddress) (UserBalanceResponse, error) {
	resp := UserBalanceResponse{
		chainId: chainId,
		user:    user,
		block:   v.Models.LatestBlock(chainId),
		tokens:  make([]types.UserBalance, 0),
	}

	balances := v.Models.RetrieveUserBalances(chainId, user)
	for _, bal := range balances {
		resp.tokens = append(resp.tokens, bal)
	}

	return resp, nil
}
