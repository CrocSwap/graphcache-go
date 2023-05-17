package views

import (
	"github.com/CrocSwap/graphcache-go/model"
	"github.com/CrocSwap/graphcache-go/types"
)

type UserPosition struct {
	types.PositionLocation
	model.PositionTracker
	types.TokenPairMetadata
}

func (v *Views) QueryUserPositions(chainId types.ChainId, user types.EthAddress) ([]UserPosition, error) {
	positions := v.Cache.RetrieveUserPositions(chainId, user)

	for key, _ := range positions {
		v.Cache.MaterializeTokenMetata(v.OnChain, chainId, key.Base)
		v.Cache.MaterializeTokenMetata(v.OnChain, chainId, key.Quote)
	}

	results := make([]UserPosition, 0)
	for key, val := range positions {
		baseMetadata := v.Cache.MaterializeTokenMetata(v.OnChain, chainId, key.Base).Poll()
		quoteMetadata := v.Cache.MaterializeTokenMetata(v.OnChain, chainId, key.Quote).Poll()
		metadata := types.PairTokenMetadata(baseMetadata, quoteMetadata)
		element := UserPosition{key, *val, metadata}
		results = append(results, element)
	}

	return results, nil
}
