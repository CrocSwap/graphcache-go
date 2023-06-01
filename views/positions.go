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
	PositionId string `json:"positionId"`
}

func (v *Views) QueryUserPositions(chainId types.ChainId, user types.EthAddress) ([]UserPosition, error) {
	positions := v.Cache.RetrieveUserPositions(chainId, user)

	results := make([]UserPosition, 0)
	for key, val := range positions {
		element := UserPosition{key, *val, formPositionId(key)}
		results = append(results, element)
	}

	sort.Sort(byTime(results))

	return results, nil
}

const MAX_POOL_POSITIONS = 100

func (v *Views) QueryPoolPositions(chainId types.ChainId,
	base types.EthAddress, quote types.EthAddress, poolIdx int, nResults int) ([]UserPosition, error) {

	loc := types.PoolLocation{
		ChainId: chainId,
		PoolIdx: poolIdx,
		Base:    base,
		Quote:   quote,
	}
	positions := v.Cache.RetrievePoolPositions(loc)

	results := make([]UserPosition, 0)
	for key, val := range positions {
		element := UserPosition{key, *val, formPositionId(key)}
		results = append(results, element)
	}

	sort.Sort(byTime(results))

	if nResults > MAX_POOL_POSITIONS {
		return make([]UserPosition, 0), fmt.Errorf("n must be below %d", MAX_POOL_POSITIONS)
	} else {
		if len(results) < nResults {
			return results, nil
		} else {
			return results[0:nResults], nil
		}
	}
}

func formPositionId(loc types.PositionLocation) string {
	hash := md5.Sum(structhash.Dump(loc, 1))
	return fmt.Sprintf("pos_%s", hex.EncodeToString(hash[:]))
}

type byTime []UserPosition

func (a byTime) Len() int           { return len(a) }
func (a byTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byTime) Less(i, j int) bool { return a[i].LatestUpdateTime > a[j].LatestUpdateTime }
