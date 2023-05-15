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
