package views

import (
	"github.com/CrocSwap/graphcache-go/model"
	"github.com/CrocSwap/graphcache-go/types"
)

func (v *Views) QueryPoolStats(chainId types.ChainId,
	base types.EthAddress, quote types.EthAddress, poolIdx int) model.AccumPoolStats {

	loc := types.PoolLocation{
		ChainId: chainId,
		PoolIdx: poolIdx,
		Base:    base,
		Quote:   quote,
	}
	return v.Cache.RetrievePoolAccum(loc)
}

func (v *Views) QueryPoolStatsFrom(chainId types.ChainId,
	base types.EthAddress, quote types.EthAddress, poolIdx int, histTime int) model.AccumPoolStats {

	loc := types.PoolLocation{
		ChainId: chainId,
		PoolIdx: poolIdx,
		Base:    base,
		Quote:   quote,
	}
	return v.Cache.RetrievePoolAccumBefore(loc, histTime)
}
