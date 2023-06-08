package views

import (
	"sort"

	"github.com/CrocSwap/graphcache-go/model"
	"github.com/CrocSwap/graphcache-go/types"
)

type PoolLiqCurve struct {
	AmbientLiq float64                `json:"ambientLiq"`
	Bumps      []*model.LiquidityBump `json:"liquidityBumps"`
}

func (v *Views) QueryPoolLiquidityCurve(chainId types.ChainId,
	base types.EthAddress, quote types.EthAddress, poolIdx int) PoolLiqCurve {

	loc := types.PoolLocation{
		ChainId: chainId,
		PoolIdx: poolIdx,
		Base:    base,
		Quote:   quote,
	}
	ambient, bumps := v.Cache.RetrievePoolLiqCurve(loc)
	sort.Sort(byTick(bumps))

	return PoolLiqCurve{
		AmbientLiq: ambient,
		Bumps:      bumps,
	}
}

type byTick []*model.LiquidityBump

func (a byTick) Len() int      { return len(a) }
func (a byTick) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

func (a byTick) Less(i, j int) bool {
	return a[i].Tick < a[j].Tick
}
