package views

import (
	"encoding/hex"
	"sort"

	"github.com/CrocSwap/graphcache-go/model"
	"github.com/CrocSwap/graphcache-go/types"
)

type UserPosition struct {
	types.PositionLocation
	model.PositionTracker
	model.APRCalcResult
	PositionId string `json:"positionId"`
}

func (v *Views) QueryUserPositions(chainId types.ChainId, user types.EthAddress) []UserPosition {
	positions := v.Cache.RetrieveUserPositions(chainId, user)

	results := make([]UserPosition, 0)
	for key, val := range positions {
		element := UserPosition{key, *val, val.CalcAPR(key), formPositionId(key)}
		results = append(results, element)
	}

	sort.Sort(byTime(results))

	return results
}

func (v *Views) QueryPoolPositions(chainId types.ChainId,
	base types.EthAddress, quote types.EthAddress, poolIdx int, nResults int,
	omitEmpty bool) []UserPosition {

	loc := types.PoolLocation{
		ChainId: chainId,
		PoolIdx: poolIdx,
		Base:    base,
		Quote:   quote,
	}

	// Retrieve X times the number of results to make it likely we have enough after filtering empty
	const EMPTY_MULT = 5

	positions := v.Cache.RetriveLastNPoolPos(loc, nResults*EMPTY_MULT)

	hasSeen := make(map[types.PositionLocation]bool, 0)
	results := make([]UserPosition, 0)

	for _, val := range positions {
		if !hasSeen[val.Loc] {
			hasSeen[val.Loc] = true

			if !omitEmpty || !val.Pos.PositionLiquidity.IsEmpty() {
				element := UserPosition{PositionLocation: val.Loc, PositionTracker: *val.Pos,
					APRCalcResult: val.Pos.CalcAPR(val.Loc), PositionId: formPositionId(val.Loc)}
				results = append(results, element)
			}
		}
	}

	sort.Sort(byTime(results))

	if len(results) > nResults {
		return results[0:nResults]
	} else {
		return results
	}
}

func (v *Views) QueryPoolApyLeaders(chainId types.ChainId,
	base types.EthAddress, quote types.EthAddress, poolIdx int, nResults int,
	omitEmpty bool) []UserPosition {

	const LAST_N_ELIGIBLE = 2000

	results := v.QueryPoolPositions(chainId, base, quote, poolIdx, LAST_N_ELIGIBLE, true)

	sort.Sort(byApr(results))

	if len(results) < nResults {
		return results
	} else {
		return results[0:nResults]
	}
}

func (v *Views) QueryUserPoolPositions(chainId types.ChainId, user types.EthAddress,
	base types.EthAddress, quote types.EthAddress, poolIdx int) []UserPosition {

	loc := types.PoolLocation{
		ChainId: chainId,
		PoolIdx: poolIdx,
		Base:    base,
		Quote:   quote,
	}
	positions := v.Cache.RetrieveUserPoolPositions(user, loc)

	results := make([]UserPosition, 0)
	for key, val := range positions {
		element := UserPosition{key, *val, val.CalcAPR(key), formPositionId(key)}
		results = append(results, element)
	}

	sort.Sort(byTime(results))

	return results
}

func (v *Views) QuerySinglePosition(chainId types.ChainId, user types.EthAddress,
	base types.EthAddress, quote types.EthAddress, poolIdx int, bidTick int, askTick int) *UserPosition {

	entries := v.QueryUserPoolPositions(chainId, user, base, quote, poolIdx)

	for _, pos := range entries {
		if pos.BidTick == bidTick && pos.AskTick == askTick {
			return &pos
		}
	}

	return nil
}

func formPositionId(loc types.PositionLocation) string {
	hash := loc.Hash()
	return "pos_" + hex.EncodeToString(hash[:])
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

type byApr []UserPosition

func (a byApr) Len() int      { return len(a) }
func (a byApr) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

func (a byApr) Less(i, j int) bool {
	// Break ties by time, the hash
	if a[i].APRCalcResult.Apr != a[j].APRCalcResult.Apr {
		return a[i].APRCalcResult.Apr > a[j].APRCalcResult.Apr
	} else if a[i].LatestUpdateTime != a[j].LatestUpdateTime {
		return a[i].LatestUpdateTime > a[j].LatestUpdateTime
	} else {
		return a[i].FirstMintTx > a[j].FirstMintTx
	}
}
