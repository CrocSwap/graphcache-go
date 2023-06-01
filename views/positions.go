package views

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"sort"

	"github.com/cnf/structhash"

	"github.com/CrocSwap/graphcache-go/model"
	"github.com/CrocSwap/graphcache-go/types"
)

type UserPosition struct {
	types.PositionLocation
	model.PositionTracker
	types.TokenPairMetadata
	PositionId string `json:"positionId"`
}

func (v *Views) QueryUserPositions(chainId types.ChainId, user types.EthAddress) ([]UserPosition, error) {
	positions := v.Cache.RetrieveUserPositions(chainId, user)

	for key := range positions {
		v.Cache.MaterializeTokenMetata(v.OnChain, chainId, key.Base)
		v.Cache.MaterializeTokenMetata(v.OnChain, chainId, key.Quote)
	}

	results := make([]UserPosition, 0)
	for key, val := range positions {
		baseMetadata := v.Cache.MaterializeTokenMetata(v.OnChain, chainId, key.Base).Poll()
		quoteMetadata := v.Cache.MaterializeTokenMetata(v.OnChain, chainId, key.Quote).Poll()
		metadata := types.PairTokenMetadata(baseMetadata, quoteMetadata)
		element := UserPosition{key, *val, metadata, formPositionId(key)}
		results = append(results, element)
	}

	sort.Sort(byTime(results))
	fmt.Println(results[0].LatestUpdateTime)

	return results, nil
}

func formPositionId(loc types.PositionLocation) string {
	hash := md5.Sum(structhash.Dump(loc, 1))
	return fmt.Sprintf("pos_%s", hex.EncodeToString(hash[:]))
}

type byTime []UserPosition

func (a byTime) Len() int           { return len(a) }
func (a byTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byTime) Less(i, j int) bool { return a[i].LatestUpdateTime > a[j].LatestUpdateTime }
