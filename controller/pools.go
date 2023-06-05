package controller

import (
	"sort"

	"github.com/CrocSwap/graphcache-go/model"
	"github.com/CrocSwap/graphcache-go/tables"
	"github.com/CrocSwap/graphcache-go/types"
)

const N_POSITIONS_REFRESH_ON_SWAP = 30

func (c *ControllerOverNetwork) resyncPoolOnSwap(l tables.Swap) []posImpactMsg {
	loc := types.PoolLocation{
		ChainId: c.chainId,
		PoolIdx: l.PoolIdx,
		Base:    types.RequireEthAddr(l.Base),
		Quote:   types.RequireEthAddr(l.Quote),
	}
	positions := c.ctrl.cache.RetrievePoolPositions(loc)

	hotPos := c.subsetRecent(positions)
	msgs := make([]posImpactMsg, 0)
	for loc, pos := range hotPos {
		msgs = append(msgs, posImpactMsg{loc, pos, l.Time})
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

type posByTime []*model.PositionTracker

func (a posByTime) Len() int      { return len(a) }
func (a posByTime) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a posByTime) Less(i, j int) bool {
	return a[i].LatestUpdateTime > a[j].LatestUpdateTime
}
