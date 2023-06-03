package views

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"sort"

	"github.com/cnf/structhash"

	"github.com/CrocSwap/graphcache-go/model"
	"github.com/CrocSwap/graphcache-go/types"
)

type UserLimitOrder struct {
	types.KOClaimLocation
	model.PositionLiquidity
	userLimitExtras
}

type userLimitExtras struct {
	LimitId          string  `json:"limitOrderId"`
	ClaimableLiq     big.Int `json:"claimableLiq"`
	CrossTime        int     `json:"crossTime"`
	LatestUpdateTime int     `json:"latestUpdateTime"`
	TimeFirstMint    int     `json:"timeFirstMint"`
}

func (v *Views) QueryUserLimits(chainId types.ChainId, user types.EthAddress) ([]UserLimitOrder, error) {
	results := make([]UserLimitOrder, 0)

	subplots := v.Cache.RetrieveUserLimits(chainId, user)
	for pos, subplot := range subplots {
		results = append(results, unrollSubplot(pos, subplot)...)
	}

	sort.Sort(byTimeLO(results))
	return results, nil
}

func (v *Views) QueryPoolLimits(chainId types.ChainId,
	base types.EthAddress, quote types.EthAddress, poolIdx int, nResults int) ([]UserLimitOrder, error) {
	results := make([]UserLimitOrder, 0)

	loc := types.PoolLocation{
		ChainId: chainId,
		PoolIdx: poolIdx,
		Base:    base,
		Quote:   quote,
	}

	for pos, subplot := range v.Cache.RetrievePoolLimits(loc) {
		results = append(results, unrollSubplot(pos, subplot)...)
	}

	sort.Sort(byTimeLO(results))

	if nResults > MAX_POOL_POSITIONS {
		return make([]UserLimitOrder, 0), fmt.Errorf("n must be below %d", MAX_POOL_POSITIONS)
	} else {
		if len(results) < nResults {
			return results, nil
		} else {
			return results[0:nResults], nil
		}
	}
}

func (v *Views) QueryUserPoolLimits(chainId types.ChainId, user types.EthAddress,
	base types.EthAddress, quote types.EthAddress, poolIdx int) ([]UserLimitOrder, error) {
	results := make([]UserLimitOrder, 0)

	loc := types.PoolLocation{
		ChainId: chainId,
		PoolIdx: poolIdx,
		Base:    base,
		Quote:   quote,
	}
	positions := v.Cache.RetrieveUserPoolLimits(user, loc)

	for pos, subplot := range positions {
		results = append(results, unrollSubplot(pos, subplot)...)
	}

	sort.Sort(byTimeLO(results))
	return results, nil
}

func (v *Views) QuerySingleLimit(chainId types.ChainId, user types.EthAddress,
	base types.EthAddress, quote types.EthAddress, poolIdx int,
	bidTick int, askTick int, isBid bool, pivotTime int) (*UserLimitOrder, error) {

	entries, err := v.QueryUserPoolLimits(chainId, user, base, quote, poolIdx)
	if err != nil {
		return nil, err
	}

	for _, pos := range entries {
		if pos.BidTick == bidTick && pos.AskTick == askTick && pos.IsBid == isBid && pos.PivotTime == pivotTime {
			return &pos, nil
		}
	}

	return nil, nil
}

func unrollSubplot(pos types.PositionLocation, subplot *model.KnockoutSubplot) []UserLimitOrder {
	unrolled := make([]UserLimitOrder, 0)

	if !subplot.Liq.Active.IsEmpty() {
		claimLoc := pos.ToClaimLoc(0)
		unrolled = append(unrolled, UserLimitOrder{
			claimLoc,
			subplot.Liq.Active,
			userLimitExtras{
				LimitId:          formLimitId(claimLoc),
				LatestUpdateTime: subplot.LatestTime,
				TimeFirstMint:    subplot.Liq.TimeFirstMint,
			}})
	}

	for pivotTime, claim := range subplot.Liq.KnockedOut {
		if !claim.IsEmpty() {
			claimLoc := pos.ToClaimLoc(pivotTime)

			crossTime, ok := subplot.GetCrossForPivotTime(pivotTime)
			if !ok {
				log.Fatalf("PivotTime=%d missing cross time", pivotTime)
			}

			unrolled = append(unrolled, UserLimitOrder{
				claimLoc,
				model.PositionLiquidity{},
				userLimitExtras{
					LimitId:          formLimitId(claimLoc),
					CrossTime:        crossTime,
					ClaimableLiq:     claim.ConcLiq,
					LatestUpdateTime: subplot.LatestTime,
					TimeFirstMint:    subplot.Liq.TimeFirstMint,
				}})
		}
	}

	return unrolled
}

func formLimitId(loc types.KOClaimLocation) string {
	hash := sha256.Sum256(structhash.Dump(loc, 1))
	return fmt.Sprintf("limit_%s", hex.EncodeToString(hash[:]))
}

type byTimeLO []UserLimitOrder

func (a byTimeLO) Len() int      { return len(a) }
func (a byTimeLO) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

func (a byTimeLO) Less(i, j int) bool {
	// Break ties by unique hash
	if a[i].LatestUpdateTime == a[j].LatestUpdateTime {
		return formLimitId(a[i].KOClaimLocation) > formLimitId(a[j].KOClaimLocation)
	}

	return a[i].LatestUpdateTime > a[j].LatestUpdateTime
}
