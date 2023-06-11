package views

import (
	"sort"

	"github.com/CrocSwap/graphcache-go/types"
)

type TokenDexAgg struct {
	TokenAddr  types.EthAddress `json:"tokenAddr"`
	DexVolume  float64          `json:"dexVolume"`
	DexFees    float64          `json:"dexFees"`
	DexTvl     float64          `json:"dexTvl"`
	LatestTime int              `json:"latestTime"`
}

func (v *Views) QueryChainStats(chainId types.ChainId, nResults int) []TokenDexAgg {

	poolAggs := v.Cache.RetrieveChainAccums(chainId)
	basin := make(map[types.EthAddress]*TokenDexAgg, 0)

	for _, poolAgg := range poolAggs {
		baseAgg, baseOk := basin[poolAgg.Base]
		quoteAgg, quoteOk := basin[poolAgg.Quote]

		if !baseOk {
			baseAgg = &TokenDexAgg{TokenAddr: poolAgg.Base}
			basin[poolAgg.Base] = baseAgg
		}
		if !quoteOk {
			quoteAgg = &TokenDexAgg{TokenAddr: poolAgg.Quote}
			basin[poolAgg.Quote] = quoteAgg
		}

		if baseAgg.LatestTime < poolAgg.LatestTime {
			baseAgg.LatestTime = poolAgg.LatestTime
		}
		if quoteAgg.LatestTime < poolAgg.LatestTime {
			quoteAgg.LatestTime = poolAgg.LatestTime
		}

		baseAgg.DexTvl += poolAgg.BaseTvl
		quoteAgg.DexTvl += poolAgg.QuoteTvl
		baseAgg.DexVolume += poolAgg.BaseVolume
		quoteAgg.DexVolume += poolAgg.QuoteVolume
		baseAgg.DexFees += poolAgg.BaseFees
		quoteAgg.DexFees += poolAgg.QuoteFees
	}

	collected := make([]TokenDexAgg, 0)
	for _, tokenAgg := range basin {
		collected = append(collected, *tokenAgg)
	}

	sort.Sort(byDexAggTime(collected))

	if len(collected) > nResults {
		return collected[:nResults]
	} else {
		return collected
	}
}

type byDexAggTime []TokenDexAgg

func (a byDexAggTime) Len() int      { return len(a) }
func (a byDexAggTime) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

func (a byDexAggTime) Less(i, j int) bool {
	return a[i].LatestTime > a[j].LatestTime
}
