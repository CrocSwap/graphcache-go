package controller

import (
	"sort"

	"github.com/CrocSwap/graphcache-go/model"
	"github.com/CrocSwap/graphcache-go/tables"
	"github.com/CrocSwap/graphcache-go/types"
)

const N_POSITIONS_REFRESH_ON_SWAP = 50

func (c *ControllerOverNetwork) resyncPoolOnSwap(l tables.Swap) ([]posImpactMsg, []koImpactMsg) {
	loc := types.PoolLocation{
		ChainId: c.chainId,
		PoolIdx: l.PoolIdx,
		Base:    types.RequireEthAddr(l.Base),
		Quote:   types.RequireEthAddr(l.Quote),
	}
	positions := c.ctrl.cache.RetrievePoolPositions(loc)
	knockouts := c.ctrl.cache.RetrievePoolLimits(loc)

	hotPos := c.subsetRecent(positions, positions)
	var msgs []posImpactMsg
	for loc, pos := range hotPos {
		msgs = append(msgs, posImpactMsg{loc, pos, l.Time})
	}

	hotKo := c.subsetRecent(positions, knockouts)
	var koMsgs []koImpactMsg
	for loc, pos := range hotKo {
		msgs = append(msgs, koImpactMsg{loc, pos, l.Time})
	}
	return msgs, koMsgs
}

type posMap = map[types.PositionLocation]*model.PositionTracker
type koMap = map[types.PositionLocation]*model.KnockoutSubplot

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

func (c *ControllerOverNetwork) subsetRecentLimits(koPositions koMap) koMap {
	var times []int
	for _, ko := range koPositions {
		if !ko.IsActiveEmpty() {
			times = append(times, ko.LatestTime)
		}
	}
	sort.Sort(sort.Reverse(sort.IntSlice(times)))

	minTime := 0
	if len(times) > N_POSITIONS_REFRESH_ON_SWAP {
		minTime = times[N_POSITIONS_REFRESH_ON_SWAP]
	}

	ret := make(koMap, 0)
	for loc, ko := range koPositions {
		if !ko.IsActiveEmpty() && ko.LatestTime >= minTime {
			ret[loc] = ko
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
