package controller

import (
	"sort"

	"github.com/CrocSwap/graphcache-go/model"
	"github.com/CrocSwap/graphcache-go/tables"
	"github.com/CrocSwap/graphcache-go/types"
)

const N_POSITIONS_REFRESH_ON_SWAP = 50

func (c *ControllerOverNetwork) resyncPoolOnSwap(l tables.Swap) []posImpactMsg {
	var msgs []posImpactMsg

	if isRecentEvent(l.Time) {
		loc := types.PoolLocation{
			ChainId: c.chainId,
			PoolIdx: l.PoolIdx,
			Base:    types.RequireEthAddr(l.Base),
			Quote:   types.RequireEthAddr(l.Quote),
		}
		positions, lock := c.ctrl.cache.BorrowPoolPositions(loc)

		hotPos := c.subsetRecent(positions)
		for loc, pos := range hotPos {
			msgs = append(msgs, posImpactMsg{loc, pos, l.Time})
		}

		if lock != nil {
			lock.RUnlock()
		}
	}
	return msgs
}

type posMap = map[types.PositionLocation]*model.PositionTracker

func (c *ControllerOverNetwork) subsetRecent(poolPositions posMap) posMap {
	var times []int
	for _, pos := range poolPositions {
		if !pos.IsEmpty() {
			times = append(times, pos.LatestUpdateTime)
		}
	}
	sort.Sort(sort.Reverse(sort.IntSlice(times)))

	minTime := 0
	if len(times) > N_POSITIONS_REFRESH_ON_SWAP {
		minTime = times[N_POSITIONS_REFRESH_ON_SWAP]
	}

	ret := make(posMap, 0)
	for loc, pos := range poolPositions {
		if !pos.IsEmpty() && pos.LatestUpdateTime >= minTime {
			ret[loc] = pos
		}
	}

	return ret
}
