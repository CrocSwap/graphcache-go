package views

import (
	"github.com/CrocSwap/graphcache-go/models"
	"github.com/CrocSwap/graphcache-go/types"
)

type UserPosition struct {
	types.PositionLocation
	models.PositionTracker
}

func (v *Views) QueryUserPositions(chainId types.ChainId, user types.EthAddress) ([]UserPosition, error) {
	positions := v.Models.RetrieveUserPositions(chainId, user)
	results := make([]UserPosition, 0)
	for key, val := range positions {
		element := UserPosition{key, *val}
		results = append(results, element)
	}
	return results, nil
}
