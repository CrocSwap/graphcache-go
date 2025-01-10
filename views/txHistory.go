package views

import (
	"encoding/hex"
	"sort"

	"github.com/CrocSwap/graphcache-go/tables"
	"github.com/CrocSwap/graphcache-go/types"
)

type UserTxHistory struct {
	types.PoolTxEvent
	EventId string `json:"txId"`
}

func (v *Views) QueryUserTxHist(chainId types.ChainId, user types.EthAddress, nResults int, afterTime int, beforeTime int) []UserTxHistory {
	var results []types.PoolTxEvent
	if afterTime == 0 && beforeTime == 0 {
		results = v.Cache.RetrieveLastNUserTxs(chainId, user, nResults)
	} else {
		results = v.Cache.RetrieveUserTxsAtTime(chainId, user, afterTime, beforeTime, nResults)
	}
	return appendTags(results)
}

func (v *Views) QueryUserPoolTxHist(chainId types.ChainId, user types.EthAddress, base types.EthAddress, quote types.EthAddress, poolIdx int, nResults int, afterTime int, beforeTime int) []UserTxHistory {
	results := v.Cache.RetrieveLastNUserTxs(chainId, user, 99999999)
	var filteredResults []types.PoolTxEvent
	loc := types.PoolLocation{
		ChainId: chainId,
		PoolIdx: poolIdx,
		Base:    base,
		Quote:   quote,
	}
	for _, tx := range results {
		if tx.PoolLocation == loc && ((beforeTime == 0 && afterTime == 0) || (tx.TxTime >= afterTime && tx.TxTime < beforeTime)) {
			filteredResults = append(filteredResults, tx)
		}
	}
	sort.Sort(byTimeTx(filteredResults))

	if len(filteredResults) > nResults {
		filteredResults = filteredResults[:nResults]
	}
	return appendTags(filteredResults)
}

func (v *Views) QueryPoolTxHist(chainId types.ChainId,
	base types.EthAddress, quote types.EthAddress, poolIdx int, nResults int, afterTime int, beforeTime int) []UserTxHistory {

	loc := types.PoolLocation{
		ChainId: chainId,
		PoolIdx: poolIdx,
		Base:    base,
		Quote:   quote,
	}
	var results []types.PoolTxEvent
	if afterTime == 0 && beforeTime == 0 {
		results = v.Cache.RetrieveLastNPoolTxs(loc, nResults)
	} else {
		results = v.Cache.RetrievePoolTxsAtTime(loc, afterTime, beforeTime, nResults)
	}
	return appendTags(results)
}

type PlumeTaskStatus struct {
	Completed *bool  `json:"completed,omitempty"`
	Error     string `json:"error,omitempty"`
	Code      int    `json:"code"`
}

func (v *Views) QueryPlumeUserTask(user types.EthAddress, task string) (status PlumeTaskStatus) {
	userTxs := v.Cache.RetrieveLastNUserTxs("0x18231", user, 99999999)
	completed := false
	switch task {
	case "ambient_deposit":
		for _, tx := range userTxs {
			if tx.ChangeType == tables.ChangeTypeMint {
				completed = true
				break
			}
		}
	case "ambient_swap":
		for _, tx := range userTxs {
			if tx.ChangeType == tables.ChangeTypeSwap {
				completed = true
				break
			}
		}
	case "ambient_limit":
		for _, tx := range userTxs {
			if tx.ChangeType == tables.ChangeTypeMint && tx.EntityType == tables.EntityTypeLimit {
				completed = true
				break
			}
		}
	default:
		status.Error = "Task is not supported"
		status.Code = 1
	}
	if status.Error == "" {
		status.Completed = &completed
	}
	return
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
	hash := loc.Hash(nil)
	return "tx_" + hex.EncodeToString(hash[:])
}

type byTimeTx []types.PoolTxEvent

func (a byTimeTx) Len() int      { return len(a) }
func (a byTimeTx) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

func (a byTimeTx) Less(i, j int) bool {
	if a[i].TxTime != a[j].TxTime {
		return a[i].TxTime > a[j].TxTime
	}

	if a[i].CallIndex != a[j].CallIndex {
		return a[i].CallIndex > a[j].CallIndex
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
