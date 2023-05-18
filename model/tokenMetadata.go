package model

import (
	"log"

	"github.com/CrocSwap/graphcache-go/loader"
	"github.com/CrocSwap/graphcache-go/types"
)

func InitTokenMetadata(onChain *loader.OnChainLoader, chain types.ChainId, token types.EthAddress) *ExpiryHandle[types.TokenMetadata] {
	fetchFn := func() types.TokenMetadata {
		metadata, err := onChain.FetchTokenMetadata(chain, token)

		if err != nil {
			log.Println("Warning: Unable to load token metadata for " + token)
		}
		return metadata
	}

	return InitCacheHandle(fetchFn, NEVER_EXPIRE_TIMEOUT, types.TokenMetadata{})
}

const NEVER_EXPIRE_TIMEOUT = 1000000
