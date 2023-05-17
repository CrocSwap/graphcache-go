package model

import (
	"log"
	"sync"

	"github.com/CrocSwap/graphcache-go/loader"
	"github.com/CrocSwap/graphcache-go/types"
)

type TokenMetadataHandle struct {
	metadata types.TokenMetadata
	wg       sync.WaitGroup
}

func InitTokenMetadata(onChain *loader.OnChainLoader, chain types.ChainId, token types.EthAddress) *TokenMetadataHandle {
	hndl := TokenMetadataHandle{}
	hndl.wg.Add(1)

	go func() {
		defer hndl.wg.Done()
		metadata, err := onChain.FetchTokenMetadata(chain, token)

		if err != nil {
			log.Println("Warning: Unable to load token metadata for " + token)
		} else {
			hndl.metadata = metadata
		}
	}()

	return &hndl
}

func (hndl *TokenMetadataHandle) Poll() types.TokenMetadata {
	hndl.wg.Wait()
	return hndl.metadata
}
