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
	positions, lock := v.Cache.BorrowPoolPositions(loc)

	results := make([]UserPosition, 0)

	for key, val := range positions {
		if !omitEmpty || !val.PositionLiquidity.IsEmpty() {
			// don't fill all the fields before sorting and truncating
			element := UserPosition{PositionLocation: key, PositionTracker: *val,
				APRCalcResult: model.APRCalcResult{0, 0, 0, 0}, PositionId: ""}

			results = append(results, element)
		}
	}

	if lock != nil {
		lock.Unlock()
	}

	sort.Sort(byTime(results))

	if len(results) > nResults {
		results = results[0:nResults]
	}

	for i, pos := range results {
		pos.APRCalcResult = pos.CalcAPR(pos.PositionLocation)
		pos.PositionId = formPositionId(pos.PositionLocation)
		results[i] = pos
	}
	return results
}

func (v *Views) QueryPoolApyLeaders(chainId types.ChainId,
	base types.EthAddress, quote types.EthAddress, poolIdx int, nResults int,
	omitEmpty bool) []UserPosition {
	results := v.QueryPoolPositions(chainId, base, quote, poolIdx, 1000000, true)

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
