package views

import (
	"crypto/sha256"
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
	base types.EthAddress, quote types.EthAddress, poolIdx int, nResults int,
	omitEmpty bool) ([]UserPosition, error) {

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
		if !omitEmpty || !val.PositionLiquidity.IsEmpty() {
			results = append(results, element)
		} else {
			fmt.Println("Omit empty")
		}
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

func (v *Views) QueryUserPoolPositions(chainId types.ChainId, user types.EthAddress,
	base types.EthAddress, quote types.EthAddress, poolIdx int) ([]UserPosition, error) {

	loc := types.PoolLocation{
		ChainId: chainId,
		PoolIdx: poolIdx,
		Base:    base,
		Quote:   quote,
	}
	positions := v.Cache.RetrieveUserPoolPositions(user, loc)

	results := make([]UserPosition, 0)
	for key, val := range positions {
		element := UserPosition{key, *val, formPositionId(key)}
		results = append(results, element)
	}

	sort.Sort(byTime(results))

	return results, nil
}

func (v *Views) QuerySinglePosition(chainId types.ChainId, user types.EthAddress,
	base types.EthAddress, quote types.EthAddress, poolIdx int, bidTick int, askTick int) (*UserPosition, error) {

	entries, err := v.QueryUserPoolPositions(chainId, user, base, quote, poolIdx)
	if err != nil {
		return nil, err
	}

	for _, pos := range entries {
		if pos.BidTick == bidTick && pos.AskTick == askTick {
			return &pos, nil
		}
	}

	return nil, nil
}

func formPositionId(loc types.PositionLocation) string {
	hash := sha256.Sum256(structhash.Dump(loc, 1))
	return fmt.Sprintf("pos_%s", hex.EncodeToString(hash[:]))
}

type byTime []UserPosition

func (a byTime) Len() int      { return len(a) }
func (a byTime) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

func (a byTime) Less(i, j int) bool {
	// Break ties by unique hash
	if a[i].LatestUpdateTime == a[j].LatestUpdateTime {
		return a[i].FirstMintTx > a[j].FirstMintTx
	}

	return a[i].LatestUpdateTime > a[j].LatestUpdateTime
}
