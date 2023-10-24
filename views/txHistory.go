package views

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"

	"github.com/CrocSwap/graphcache-go/types"
	"github.com/cnf/structhash"
)

type UserTxHistory struct {
	types.PoolTxEvent
	EventId string `json:"txId"`
}

func (v *Views) QueryUserTxHist(chainId types.ChainId, user types.EthAddress, nResults int) []UserTxHistory {
	results := v.Cache.RetrieveUserTxs(chainId, user)
	sort.Sort(byTimeTx(results))
	if len(results) < nResults {
		return appendTags(results)
	} else {
		return appendTags(results[0:nResults])
	}
}

func (v *Views) QueryPoolTxHist(chainId types.ChainId,
	base types.EthAddress, quote types.EthAddress, poolIdx int, nResults int) []UserTxHistory {

	loc := types.PoolLocation{
		ChainId: chainId,
		PoolIdx: poolIdx,
		Base:    base,
		Quote:   quote,
	}
	results := v.Cache.RetrievePoolTxs(loc)
	sort.Sort(byTimeTx(results))

	if len(results) < nResults {
		return appendTags(results)
	} else {
		return appendTags(results[0:nResults])
	}
}

func (v *Views) QueryPoolTxHistFrom(chainId types.ChainId,
	base types.EthAddress, quote types.EthAddress, poolIdx int, nResults int,
	time int, period int) []UserTxHistory {
	txs := v.QueryPoolTxHist(chainId, base, quote, poolIdx, 1000000)

	var results []UserTxHistory
	for _, tx := range txs {
		if tx.TxTime >= time && tx.TxTime < time+period {
			results = append(results, tx)
		}
	}

	if len(results) < nResults {
		return results
	} else {
		return results[0:nResults]
	}
}

func appendTags(txs []types.PoolTxEvent) []UserTxHistory {
	var results []UserTxHistory
	for _, tx := range txs {
		entry := UserTxHistory{
			tx,
			formTxId(tx),
		}
		results = append(results, entry)
	}
	return results
}

func formTxId(loc types.PoolTxEvent) string {
	hash := sha256.Sum256(structhash.Dump(loc, 1))
	return fmt.Sprintf("tx_%s", hex.EncodeToString(hash[:]))
}

type byTimeTx []types.PoolTxEvent

func (a byTimeTx) Len() int      { return len(a) }
func (a byTimeTx) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

func (a byTimeTx) Less(i, j int) bool {
	if a[i].TxTime != a[j].TxTime {
		return a[i].TxTime > a[j].TxTime
	}

	// Tie breakers if occurs at same time
	if a[i].ChangeType != a[j].ChangeType {
		return a[i].ChangeType > a[j].ChangeType
	}

	if a[i].PositionType != a[j].PositionType {
		return a[i].PositionType > a[j].PositionType
	}

	if string(a[i].Base) != string(a[j].Base) {
		return a[i].Base > a[j].Base
	}

	if string(a[i].Quote) != string(a[j].Quote) {
		return a[i].Quote > a[j].Quote
	}

	if a[i].BidTick != a[j].BidTick {
		return a[i].BidTick > a[j].BidTick
	}

	if a[i].AskTick != a[j].AskTick {
		return a[i].BidTick > a[j].BidTick
	}

	return false
}
