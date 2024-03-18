package controller

import (
	"github.com/CrocSwap/graphcache-go/tables"
	"github.com/CrocSwap/graphcache-go/types"
)

const N_POSITIONS_REFRESH_ON_SWAP = 100

func (c *ControllerOverNetwork) resyncPoolOnSwap(l tables.Swap) []posImpactMsg {
	var msgs []posImpactMsg

	if isRecentEvent(l.Time) {
		loc := types.PoolLocation{
			ChainId: c.chainId,
			PoolIdx: l.PoolIdx,
			Base:    types.RequireEthAddr(l.Base),
			Quote:   types.RequireEthAddr(l.Quote),
		}
		positions := c.ctrl.cache.RetriveLastNPoolPos(loc, N_POSITIONS_REFRESH_ON_SWAP)

		for _, pos := range positions {
			if !pos.Pos.IsEmpty() {
				msgs = append(msgs, posImpactMsg{pos.Loc, pos.Pos, l.Time})
			}
		}
	}
	return msgs
}

/*type posMap = map[types.PositionLocation]*model.PositionTracker

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
}*/
