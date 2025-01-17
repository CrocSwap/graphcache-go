package views

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"sort"

	"github.com/CrocSwap/graphcache-go/cache"
	"github.com/CrocSwap/graphcache-go/model"
	"github.com/CrocSwap/graphcache-go/tables"
	"github.com/CrocSwap/graphcache-go/types"
)

type UserPosition struct {
	types.PositionLocation
	model.PositionTracker
	model.APRCalcResult
	PositionId string `json:"positionId"`
}

type HistoricUserPosition struct {
	types.PositionLocation
	PositionType  tables.PosType           `json:"positionType"`
	TimeFirstMint int                      `json:"timeFirstMint"`
	FirstMintTx   string                   `json:"firstMintTx"`
	PositionId    string                   `json:"positionId"`
	Liq           *big.Int                 `json:"liq"`
	BaseTokens    *big.Int                 `json:"baseTokens"`
	QuoteTokens   *big.Int                 `json:"quoteTokens"`
	PoolPrice     float64                  `json:"poolPrice"`
	LiqHist       model.LiquidityDeltaHist `json:"liqHist"`
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
	omitEmpty bool, afterTime int, beforeTime int) []UserPosition {
	results := make([]UserPosition, 0, nResults)

	loc := types.PoolLocation{
		ChainId: chainId,
		PoolIdx: poolIdx,
		Base:    base,
		Quote:   quote,
	}

	// Retrieve X times the number of results to make it likely we have enough after filtering empty.
	// Retrieving more than needed is faster than retrieving multiple times.
	EMPTY_MULT := 20
	if omitEmpty {
		EMPTY_MULT = 2 // >1 because we might have to skip duplicate updates
	} else if nResults > 200 { // for APY leaderboard
		EMPTY_MULT = 2
	} else if nResults < 10 { // so that small queries don't stall for a long time
		EMPTY_MULT = 50
	}
	hasSeen := make(map[[32]byte]struct{}, nResults*EMPTY_MULT*2)
	buf := new(bytes.Buffer)
	buf.Grow(300)

	// Keep retrieving updates until we have nResults or until iter limit is reached.
	// Without the limit it could stall for like 10 seconds searching for non-empty
	// positions, especially if the startup RPC liq refresh isn't done yet.
	for i := 0; i < 5; i++ {
		var positions []cache.PosAndLocPair
		if afterTime == 0 && beforeTime == 0 {
			positions = v.Cache.RetrieveLastNPoolPos(loc, nResults*EMPTY_MULT)
		} else {
			positions = v.Cache.RetrievePoolPosAtTime(loc, afterTime, beforeTime, nResults*EMPTY_MULT, hasSeen)
		}

		for _, val := range positions {
			hash := val.Loc.CachedHash
			if _, ok := hasSeen[hash]; !ok || (afterTime != 0 || beforeTime != 0) {
				hasSeen[hash] = struct{}{}

				if !omitEmpty || !val.Pos.PositionLiquidity.IsEmpty() {
					element := UserPosition{PositionLocation: val.Loc, PositionTracker: *val.Pos,
						APRCalcResult: val.Pos.CalcAPR(val.Loc), PositionId: formPositionId(val.Loc)}
					results = append(results, element)
				}
			}
			if len(results) >= nResults {
				break
			}
		}

		if len(results) >= nResults {
			results = results[0:nResults]
			break
		} else if len(positions) == 0 {
			break
		} else {
			beforeTime = positions[len(positions)-1].Time() - 1
		}
	}
	sort.Sort(byTime(results))
	return results
}

func (v *Views) QueryPoolApyLeaders(chainId types.ChainId,
	base types.EthAddress, quote types.EthAddress, poolIdx int, nResults int,
	omitEmpty bool) []UserPosition {

	const LAST_N_ELIGIBLE = 2000

	results := v.QueryPoolPositions(chainId, base, quote, poolIdx, LAST_N_ELIGIBLE, true, 0, 0)

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

func (v *Views) QueryHistoricPositions(chainId types.ChainId, base types.EthAddress, quote types.EthAddress,
	poolIdx int, time int, user types.EthAddress, omitEmpty bool) []HistoricUserPosition {
	livePositions := make([]HistoricUserPosition, 0)

	pools := v.Cache.RetrievePoolSet()
	for _, loc := range pools {
		if (chainId != "" && loc.ChainId != chainId) || (base != "" && loc.Base != base) || (quote != "" && loc.Quote != quote) || (poolIdx != 0 && loc.PoolIdx != poolIdx) {
			continue
		}
		positions := v.Cache.RetrievePoolPositions(loc)

		poolHistPos := make([]HistoricUserPosition, 0)

		for key, val := range positions {
			if user != "" && user != key.User {
				continue
			}
			histPos := HistoricUserPosition{
				PositionLocation: key,
				PositionId:       formPositionId(key),
				TimeFirstMint:    val.TimeFirstMint,
				FirstMintTx:      val.FirstMintTx,
				PositionType:     val.PositionType,
				LiqHist:          val.LiqHist,
			}
			poolHistPos = append(poolHistPos, histPos)
		}

		lastTrade, _ := v.Cache.RetrievePoolAccumBefore(loc, time)
		poolPrice := lastTrade.LastPriceIndic
		for _, position := range poolHistPos {
			var liqSum float64
			for _, liqChange := range position.LiqHist.Hist {
				if liqChange.Time <= time {
					liqSum += liqChange.LiqChange
				}
			}
			if liqSum <= 0 {
				liqSum = 0
				if omitEmpty {
					continue
				}
			}
			if position.TimeFirstMint > time {
				continue
			}
			if position.PositionType == tables.PosTypeConcentrated {
				liqSumBig, _ := big.NewFloat(liqSum).Int(nil)
				position.Liq = liqSumBig
				position.BaseTokens, position.QuoteTokens = model.DeriveTokensFromConcLiquidity(liqSum, position.BidTick, position.AskTick, poolPrice)
			} else if position.PositionType == tables.PosTypeAmbient {
				liqSumBig, _ := big.NewFloat(liqSum).Int(nil)
				position.Liq = liqSumBig
				position.BaseTokens, position.QuoteTokens = model.DeriveTokensFromAmbLiquidity(liqSum, poolPrice)
			}
			position.PoolPrice = poolPrice
			livePositions = append(livePositions, position)
		}
	}
	return livePositions
}

func formPositionId(loc types.PositionLocation) string {
	return "pos_" + hex.EncodeToString(loc.CachedHash[:])
}

type byTime []UserPosition

func (a byTime) Len() int      { return len(a) }
func (a byTime) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

func (a byTime) Less(i, j int) bool {
	// Break ties by unique hash
	if a[i].LatestUpdateTime != a[j].LatestUpdateTime {
		return a[i].LatestUpdateTime > a[j].LatestUpdateTime
	}
	if a[i].FirstMintTx != a[j].FirstMintTx {
		return a[i].FirstMintTx > a[j].FirstMintTx
	}

	concCmp := a[i].ConcLiq.Cmp(&a[j].ConcLiq)
	if concCmp != 0 {
		return concCmp > 0
	}
	return a[i].AmbientLiq.Cmp(&a[j].AmbientLiq) > 0

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
