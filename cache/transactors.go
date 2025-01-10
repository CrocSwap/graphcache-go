package cache

import (
	"log"
	"slices"
	"sync"
	"time"

	"github.com/CrocSwap/graphcache-go/model"
	"github.com/CrocSwap/graphcache-go/types"
)

func (m *MemoryCache) LatestBlock(chainId types.ChainId) int64 {
	block, okay := m.latestBlocks.lookup(chainId)
	if okay {
		return block
	} else {
		return -1
	}
}

func (m *MemoryCache) RetrieveUserBalances(chainId types.ChainId, user types.EthAddress) []types.EthAddress {
	key := chainAndAddr{chainId, user}
	tokens, _ := m.userBalTokens.lookup(key)
	return tokens
}

func (m *MemoryCache) RetrieveLastNUserTxs(chainId types.ChainId, user types.EthAddress, nResults int) []types.PoolTxEvent {
	key := chainAndAddr{chainId, user}
	txs, _ := m.userTxs.lookupLastN(key, nResults)
	return txs
}

func (m *MemoryCache) RetrieveUserTxsAtTime(chainId types.ChainId, user types.EthAddress, afterTime int, beforeTime int, nResults int) []types.PoolTxEvent {
	key := chainAndAddr{chainId, user}
	txs, _ := m.userTxs.lookupLastNAtTime(key, afterTime, beforeTime, nResults)
	return txs
}

func (m *MemoryCache) RetrievePoolSet() []types.PoolLocation {
	return m.poolTradingHistory.keySet()
}

func (m *MemoryCache) RetrievePoolTxs(pool types.PoolLocation) []types.PoolTxEvent {
	txs, _ := m.poolTxs.lookupCopy(pool)
	return txs
}

func (m *MemoryCache) RetrieveLastNPoolTxs(pool types.PoolLocation, lastN int) []types.PoolTxEvent {
	txs, _ := m.poolTxs.lookupLastN(pool, lastN)
	return txs
}

func (m *MemoryCache) RetrievePoolTxsAtTime(pool types.PoolLocation, afterTime int, beforeTime int, n int) []types.PoolTxEvent {
	txs, _ := m.poolTxs.lookupLastNAtTime(pool, afterTime, beforeTime, n)
	return txs
}

func (m *MemoryCache) RetrieveLastNPoolPos(pool types.PoolLocation, lastN int) []PosAndLocPair {
	txs, _ := m.poolPosUpdates.lookupLastN(pool, lastN)
	return txs
}

func (m *MemoryCache) RetrievePoolPosAtTime(pool types.PoolLocation, afterTime int, beforeTime int, n int, seen map[[32]byte]struct{}) []PosAndLocPair {
	txs, _ := m.poolPosUpdates.lookupLastNTimeNonUnique(pool, afterTime, beforeTime, n, seen)
	return txs
}

func (m *MemoryCache) RetrieveLastNPoolKo(pool types.PoolLocation, lastN int) []KoAndLocPair {
	txs, _ := m.poolKoUpdates.lookupLastN(pool, lastN)
	return txs
}

func (m *MemoryCache) RetrievePoolKoAtTime(pool types.PoolLocation, afterTime int, beforeTime int, n int, seen map[[32]byte]struct{}) []KoAndLocPair {
	txs, _ := m.poolKoUpdates.lookupLastNTimeNonUnique(pool, afterTime, beforeTime, n, seen)
	return txs
}

func (m *MemoryCache) RetrieveUserPositions(chainId types.ChainId, user types.EthAddress) map[types.PositionLocation]*model.PositionTracker {
	key := chainAndAddr{chainId, user}
	pos, okay := m.userPositions.lookupSet(key)
	if okay {
		return pos
	} else {
		return make(map[types.PositionLocation]*model.PositionTracker)
	}
}

func (m *MemoryCache) RetrieveAllPositions() map[types.PositionLocation]*model.PositionTracker {
	return m.liqPosition.clone()
}

func (m *MemoryCache) RetrieveAllCurves() map[types.PoolLocation]*model.LiquidityCurve {
	return m.poolLiqCurve.clone()
}

// Returns all positions sorted by LatestUpdateTime in descending order
func (m *MemoryCache) RetrieveAllPositionsSorted() []PosAndLocPair {
	allPos := m.liqPosition.clone()
	posUpdates := make([]PosAndLocPair, 0, len(allPos))
	for loc, pos := range allPos {
		posUpdates = append(posUpdates, PosAndLocPair{loc, pos})
	}
	slices.SortFunc(posUpdates, func(a, b PosAndLocPair) int {
		return b.Pos.LatestUpdateTime - a.Pos.LatestUpdateTime
	})
	return posUpdates
}

func (m *MemoryCache) RetrieveUserLimits(
	chainId types.ChainId,
	user types.EthAddress) map[types.PositionLocation]*model.KnockoutSubplot {
	key := chainAndAddr{chainId, user}
	pos, okay := m.userKnockouts.lookupSet(key)
	if okay {
		return pos
	} else {
		return make(map[types.PositionLocation]*model.KnockoutSubplot)
	}
}

func (m *MemoryCache) RetrievePoolLimits(loc types.PoolLocation) map[types.PositionLocation]*model.KnockoutSubplot {
	pos, okay := m.poolKnockouts.lookupSet(loc)
	if okay {
		return pos
	} else {
		return make(map[types.PositionLocation]*model.KnockoutSubplot)
	}
}

func (m *MemoryCache) RetrievePoolPositions(loc types.PoolLocation) map[types.PositionLocation]*model.PositionTracker {
	pos, okay := m.poolPositions.lookupSet(loc)
	if okay {
		return pos
	} else {
		return make(map[types.PositionLocation]*model.PositionTracker)
	}
}

// Caller has responsibility of unlocking the result when complete, otherwise will deadlock
func (m *MemoryCache) BorrowPoolLimits(loc types.PoolLocation) (map[types.PositionLocation]*model.KnockoutSubplot, *sync.RWMutex) {
	return m.poolKnockouts.lockSet(loc)
}

// Caller has responsibility of unlocking the result when complete, otherwise will deadlock
func (m *MemoryCache) BorrowPoolPositions(loc types.PoolLocation) (map[types.PositionLocation]*model.PositionTracker, *sync.RWMutex) {
	return m.poolPositions.lockSet(loc)
}

func (m *MemoryCache) RetrievePoolLiqCurve(loc types.PoolLocation) (float64, []*model.LiquidityBump) {
	var returnVal []*model.LiquidityBump
	ambientLiq := 0.0

	pos, okay, lock := m.poolLiqCurve.lockLookup(loc, false)
	if okay {
		defer lock.RUnlock()
		for _, bump := range pos.Bumps {
			returnVal = append(returnVal, bump)
		}
		ambientLiq = pos.AmbientLiq
	}
	return ambientLiq, returnVal
}

func (m *MemoryCache) RetrievePoolAccum(loc types.PoolLocation) (stats model.AccumPoolStats, eventCount int) {
	pos, okay := m.poolTradingHistory.lookup(loc)
	if !okay {
		return model.AccumPoolStats{}, 0
	}
	return pos.StatsCounter, len(pos.TimeSnaps) + 1
}

func (m *MemoryCache) RetrievePoolAccumFirst(loc types.PoolLocation) model.AccumPoolStats {
	pos, okay := m.poolTradingHistory.lookup(loc)
	if !okay {
		return model.AccumPoolStats{}
	}
	if len(pos.TimeSnaps) == 0 {
		return pos.StatsCounter
	} else {
		return pos.TimeSnaps[0]
	}
}

type AccumTagged struct {
	model.AccumPoolStats
	types.PoolLocation
}

func (m *MemoryCache) RetrieveChainAccums(chainId types.ChainId) []AccumTagged {
	retVal := make([]AccumTagged, 0)
	fullUniv := m.poolTradingHistory.clone()
	for loc, hist := range fullUniv {
		if loc.ChainId == chainId {
			retVal = append(retVal, AccumTagged{hist.StatsCounter, loc})
		}
	}
	return retVal
}

func (m *MemoryCache) RetrievePoolAccumBefore(loc types.PoolLocation, histTime int) (stats model.AccumPoolStats, eventCount int) {
	pos, okay, lock := m.poolTradingHistory.lockLookup(loc, false)
	if !okay {
		return model.AccumPoolStats{}, 0
	}

	defer lock.RUnlock()

	if histTime >= pos.StatsCounter.LatestTime {
		return pos.StatsCounter, len(pos.TimeSnaps) + 1
	} else {
		// Iteration in reverse order because the frontend requests only 24 hours back
		for i := len(pos.TimeSnaps) - 1; i >= 0; i-- {
			if pos.TimeSnaps[i].LatestTime <= histTime {
				return pos.TimeSnaps[i], i + 1
			}
		}
		// If histTime is before the first snapshot then return nothing
		return model.AccumPoolStats{}, 0
	}
}

func (m *MemoryCache) RetrievePoolAccumSeries(loc types.PoolLocation, startTime int, endTime int) (openVal model.AccumPoolStats, retSeries []model.AccumPoolStats) {
	retSeries = make([]model.AccumPoolStats, 0, 1000)

	start := time.Now()
	pool, okay, lock := m.poolTradingHistory.lockLookup(loc, false)
	if !okay {
		return
	}
	diff := time.Since(start)
	if diff > 200*time.Millisecond {
		log.Println("Slow lock:", diff)
	}
	defer lock.RUnlock()

	if pool.StatsCounter.LatestTime >= startTime && pool.StatsCounter.LatestTime < endTime {
		retSeries = append(retSeries, pool.StatsCounter)
	} else if len(pool.TimeSnaps) > 0 && endTime < pool.TimeSnaps[0].LatestTime {
		return
	} else if pool.StatsCounter.LatestTime < startTime {
		openVal = pool.StatsCounter
		return
	}
	start = time.Now()
	var i int
	for i = len(pool.TimeSnaps) - 1; i >= 0; i-- {
		if pool.TimeSnaps[i].LatestTime >= startTime && pool.TimeSnaps[i].LatestTime < endTime {
			retSeries = append(retSeries, pool.TimeSnaps[i])
		} else if pool.TimeSnaps[i].LatestTime < startTime {
			openVal = pool.TimeSnaps[i]
			break
		}
	}
	diff = time.Since(start)
	if diff > 200*time.Millisecond {
		log.Println("Slow loop:", diff)
	}

	start = time.Now()
	slices.Reverse(retSeries)
	diff = time.Since(start)
	if diff > 50*time.Millisecond {
		log.Println("Slow reverse:", diff)
	}
	// If entire history was added to the series, then the openVal is the first element
	if openVal.LatestTime == 0 && len(pool.TimeSnaps) > 0 {
		openVal = pool.TimeSnaps[0]
	}

	return
}

func (m *MemoryCache) RetrievePoolAccumSeriesOld(loc types.PoolLocation, startTime int, endTime int) (model.AccumPoolStats, []model.AccumPoolStats) {
	retSeries := make([]model.AccumPoolStats, 0)
	openVal, _ := m.RetrievePoolAccumBefore(loc, startTime)

	pos, okay := m.poolTradingHistory.lookup(loc)
	if !okay {
		return openVal, retSeries
	}

	defer m.poolTradingHistory.lock.RUnlock()
	m.poolTradingHistory.lock.RLock()

	for _, accum := range pos.TimeSnaps {
		if accum.LatestTime >= startTime && accum.LatestTime < endTime {
			retSeries = append(retSeries, accum)
		}
	}
	return openVal, retSeries
}

func (m *MemoryCache) BorrowPoolHourlyCandles(loc types.PoolLocation, writeLock bool) (*[]model.Candle, *sync.RWMutex) {
	candles, okay, lock := m.poolHourlyCandles.lockLookup(loc, writeLock)
	if !okay {
		candlesO := make([]model.Candle, 0)
		candles = &candlesO
		lock = m.poolHourlyCandles.insert(loc, candles)
		if writeLock {
			lock.Lock()
		} else {
			lock.RLock()
		}
	}
	return candles, lock
}

func (m *MemoryCache) RetrieveUserPoolPositions(user types.EthAddress, pool types.PoolLocation) map[types.PositionLocation]*model.PositionTracker {
	userPositions := m.RetrieveUserPositions(pool.ChainId, user)
	filtered := make(map[types.PositionLocation]*model.PositionTracker)
	for loc, pos := range userPositions {
		if loc.PoolLocation == pool {
			filtered[loc] = pos
		}
	}
	return filtered
}

func (m *MemoryCache) RetrieveUserPoolLimits(user types.EthAddress, pool types.PoolLocation) map[types.PositionLocation]*model.KnockoutSubplot {
	userLimits := m.RetrieveUserLimits(pool.ChainId, user)
	filtered := make(map[types.PositionLocation]*model.KnockoutSubplot)
	for loc, pos := range userLimits {
		if loc.PoolLocation == pool {
			filtered[loc] = pos
		}
	}
	return filtered
}

func (m *MemoryCache) AddUserBalance(chainId types.ChainId, user types.EthAddress, token types.EthAddress) {
	key := chainAndAddr{chainId, user}
	m.userBalTokens.insert(key, token)
}

func (m *MemoryCache) AddPoolEvent(tx types.PoolTxEvent) {
	userKey := chainAndAddr{tx.ChainId, tx.User}
	// m.userTxs.insert(userKey, tx)
	// m.poolTxs.insert(tx.PoolLocation, tx)
	m.userTxs.insertSorted(userKey, tx, func(i, j types.PoolTxEvent) bool {
		if i.TxTime != j.TxTime {
			return i.TxTime > j.TxTime
		}

		if i.CallIndex != j.CallIndex {
			return i.CallIndex > j.CallIndex
		}

		// Tie breakers if occurs at same time
		if i.ChangeType != j.ChangeType {
			return i.ChangeType > j.ChangeType
		}

		if i.PositionType != j.PositionType {
			return i.PositionType > j.PositionType
		}

		if i.Base != j.Base {
			return i.Base > j.Base
		}

		if i.Quote != j.Quote {
			return i.Quote > j.Quote
		}

		if i.BidTick != j.BidTick {
			return i.BidTick > j.BidTick
		}

		if i.AskTick != j.AskTick {
			return i.BidTick > j.BidTick
		}
		return false
	})
	m.poolTxs.insertSorted(tx.PoolLocation, tx, func(i, j types.PoolTxEvent) bool {
		if i.TxTime != j.TxTime {
			return i.TxTime > j.TxTime
		}

		if i.CallIndex != j.CallIndex {
			return i.CallIndex > j.CallIndex
		}

		// Tie breakers if occurs at same time
		if i.ChangeType != j.ChangeType {
			return i.ChangeType > j.ChangeType
		}

		if i.PositionType != j.PositionType {
			return i.PositionType > j.PositionType
		}

		if i.BidTick != j.BidTick {
			return i.BidTick > j.BidTick
		}

		if i.AskTick != j.AskTick {
			return i.BidTick > j.BidTick
		}
		return false
	})
}

func (m *MemoryCache) MaterializePoolLiqCurve(loc types.PoolLocation, writeLock bool) (*model.LiquidityCurve, *sync.RWMutex) {
	val, okay, lock := m.poolLiqCurve.lockLookup(loc, writeLock)
	if !okay {
		val = model.NewLiquidityCurve()
		lock = m.poolLiqCurve.insert(loc, val)
		if writeLock {
			lock.Lock()
		} else {
			lock.RLock()
		}
	}
	return val, lock
}

func (m *MemoryCache) MaterializePoolTradingHist(loc types.PoolLocation, writeLock bool) (*model.PoolTradingHistory, *sync.RWMutex) {
	val, okay, lock := m.poolTradingHistory.lockLookup(loc, writeLock)
	if !okay {
		val = model.NewPoolTradingHistory()
		lock = m.poolTradingHistory.insert(loc, val)
		if writeLock {
			lock.Lock()
		} else {
			lock.RLock()
		}
	}
	return val, lock
}

func (m *MemoryCache) BorrowPoolTradingHist(loc types.PoolLocation, writeLock bool) (*model.PoolTradingHistory, *sync.RWMutex) {
	val, _, lock := m.poolTradingHistory.lockLookup(loc, writeLock)
	return val, lock
}

func (m *MemoryCache) MaterializePosition(loc types.PositionLocation) *model.PositionTracker {
	val, ok := m.liqPosition.lookup(loc)
	if !ok {
		val = &model.PositionTracker{}
		m.liqPosition.insert(loc, val)
		m.userPositions.insert(chainAndAddr{loc.ChainId, loc.User}, loc, val)
		m.poolPositions.insert(loc.PoolLocation, loc, val)
		// m.userAndPoolPositions.insert(
		// 	chainUserAndPool{loc.User, loc.PoolLocation}, loc, val)
	}

	m.poolPosUpdates.insert(loc.PoolLocation, PosAndLocPair{loc, val})
	return val
}

func (m *MemoryCache) MaterializeKnockoutSaga(loc types.BookLocation) *model.KnockoutSaga {
	val, ok := m.knockoutSagas.lookup(loc)
	if !ok {
		val = model.NewKnockoutSaga()
		m.knockoutSagas.insert(loc, val)
	}
	return val
}

func (m *MemoryCache) MaterializeKnockoutPos(loc types.PositionLocation) *model.KnockoutSubplot {
	val, ok := m.liqKnockouts.lookup(loc)
	if !ok {
		saga := m.MaterializeKnockoutSaga(loc.ToBookLoc())
		val = saga.ForUser(loc.User)
		m.liqKnockouts.insert(loc, val)
		m.userKnockouts.insert(chainAndAddr{loc.ChainId, loc.User}, loc, val)
		m.poolKnockouts.insert(loc.PoolLocation, loc, val)
	}

	m.poolKoUpdates.insert(loc.PoolLocation, KoAndLocPair{loc, val})
	return val
}

func (m *MemoryCache) RetrievePivotTime(loc types.BookLocation) int {
	pos, okay := m.knockoutPivotTimes.lookup(loc)
	if okay {
		return pos
	} else {
		return 0
	}
}

func (m *MemoryCache) SetPivotTime(loc types.BookLocation, pivotTime int) {
	m.knockoutPivotTimes.insert(loc, pivotTime)
}
