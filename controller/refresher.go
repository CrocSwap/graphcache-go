package controller

import (
	"encoding/hex"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/CrocSwap/graphcache-go/loader"
)

type LiquidityRefresher struct {
	// Priority queue to refresh new positions immediately.
	workUrgent chan IRefreshHandle
	// Slow queue to process periodic refreshes after urgent ones.
	workSlow chan IRefreshHandle
	// Map of hndl.Hash()->urgent, to prevent duplicate refreshes. If urgent is true that means the
	// handle is either in the workUrgent queue or in both queues (in which case it will be skipped
	// if it is read from the workSlow queue). If urgent wasn't stored then there could have been a
	// situation where an urgent refresh would be ignored because there was a slow refresh pending.
	pending        map[[32]byte]bool
	pendingLock    sync.Mutex
	postProcess    chan string
	query          *loader.ICrocQuery
	lastRefreshSec int64
}

const NUM_PARALLEL_WORKERS = 200 // Should be higher than multicall_max_batch for the given chain
const URGENT_QUEUE_SIZE = 50000
const SLOW_QUEUE_SIZE = 1200000 // On Scroll about 700000 is needed for the startup refresh.
const MAX_REQS_PER_SEC = 1000

func NewLiquidityRefresher(query *loader.ICrocQuery) *LiquidityRefresher {
	liqRefresher := LiquidityRefresher{
		workUrgent:  make(chan IRefreshHandle, URGENT_QUEUE_SIZE),
		workSlow:    make(chan IRefreshHandle, SLOW_QUEUE_SIZE),
		pending:     make(map[[32]byte]bool),
		pendingLock: sync.Mutex{},
		query:       query,
		postProcess: make(chan string),
	}

	go liqRefresher.watchPostProcess()
	for idx := 0; idx < NUM_PARALLEL_WORKERS; idx += 1 {
		go liqRefresher.watchWork()
	}

	return &liqRefresher
}

/* Historical events won't require followup, because an RPC call should
 * be sync'd */
const RECENT_EVENT_WINDOW = 60
const SKIPPABLE_REFRESH_INTERVAL = 30 // Skippable requests (rewards) will only be refreshed this often per position

func isRecentEvent(eventTime int) bool {
	currentTime := time.Now().Unix()
	return int(currentTime)-eventTime < RECENT_EVENT_WINDOW
}

func (lr *LiquidityRefresher) PushRefresh(hndl IRefreshHandle, eventTime int) {
	urgent := isRecentEvent(eventTime)
	lr.requestRefresh(hndl, urgent)
	if urgent {
		go lr.pushFollowup(hndl)
	}
}

func (lr *LiquidityRefresher) PushRefreshPoll(hndl IRefreshHandle) {
	lr.requestRefresh(hndl, false)
	// Since these are periodic polls, no followup to force convergence on event
	// state is necessary
}

func (lr *LiquidityRefresher) requestRefresh(hndl IRefreshHandle, urgent bool) {
	if time.Now().Unix()-hndl.RefreshTime() < SKIPPABLE_REFRESH_INTERVAL && hndl.Skippable() {
		return
	}
	hash := hndl.Hash()
	lr.pendingLock.Lock()
	defer lr.pendingLock.Unlock()
	queuedUrgent, alreadyQueued := lr.pending[hash]
	if !alreadyQueued || (urgent && !queuedUrgent) {
		lr.pending[hash] = urgent
		if urgent {
			lr.workUrgent <- hndl
		} else {
			lr.workSlow <- hndl
		}
	}
}

const FOLLOWUP_WINDOW = 5 // Followup refreshes are grouped to be sent at this interval

/* Followup after refresh, in case RPC node hasn't synced to last
 * block at first call. */
func (lr *LiquidityRefresher) pushFollowup(hndl IRefreshHandle) {
	REFRESH_FOLLOWUP_SECS := []time.Duration{2, 10, 30, 60}
	for _, interval := range REFRESH_FOLLOWUP_SECS {
		refreshTime := time.Now().Add(interval * time.Second)
		nextWindow := (refreshTime.Unix() - refreshTime.Unix()%FOLLOWUP_WINDOW) + FOLLOWUP_WINDOW
		sleepUntilNextWindow := time.Until(time.Unix(nextWindow, 0))
		time.Sleep(sleepUntilNextWindow)
		lr.requestRefresh(hndl, true)
	}
}

const RETRY_QUERY_MIN_WAIT = 10
const RETRY_QUERY_MAX_WAIT = 60

// Should be high because temporary RPC issues should not cause a crash,
// especially since running without RPC data isn't that bad these days.
const N_MAX_RETRIES = 500

// Do this so that in case the problem is overloading the RPC, calls don't all spam again
// at same deterministic time
func retryWaitRandom() {
	waitTime := rand.Intn(RETRY_QUERY_MAX_WAIT-RETRY_QUERY_MIN_WAIT) + RETRY_QUERY_MIN_WAIT
	time.Sleep(time.Duration(waitTime) * time.Second)
}

func (lr *LiquidityRefresher) watchWork() {
	hndlPending := false
	var hndl IRefreshHandle
	defer func() {
		if r := recover(); r != nil {
			hash := hndl.Hash()
			log.Println("Panic recovered in watchWork for id", hex.EncodeToString(hash[:]), r)
			if hndlPending && hndl != nil {
				lr.workUrgent <- hndl
			}
			go lr.watchWork()
		}
	}()
	lastSec := time.Now().Unix()
	callCnt := 0
	totalCnt := 0

	for {
		fromSlowQueue := false
		// Try to read from workUrgent first, then from workSlow
		select {
		case hndl = <-lr.workUrgent:
		default:
			select {
			case hndl = <-lr.workUrgent:
			case hndl = <-lr.workSlow:
				fromSlowQueue = true
			}
		}
		// Check whether the request has been upgraded to urgent, and skip it if it was
		hash := hndl.Hash()
		lr.pendingLock.Lock()
		urgent, queued := lr.pending[hash]
		if (fromSlowQueue && urgent) || !queued {
			lr.pendingLock.Unlock()
			continue
		}
		lr.pendingLock.Unlock()

		hndlPending = true
		hndl.RefreshQuery(lr.query)
		lr.pendingLock.Lock()
		delete(lr.pending, hash)
		hndlPending = false
		lr.pendingLock.Unlock()
		lr.postProcess <- hndl.LabelTag()

		nowSec := time.Now().Unix()
		if nowSec != lastSec {
			callCnt = 0
			lastSec = nowSec

		} else {
			callCnt += 1
			if callCnt > MAX_REQS_PER_SEC {
				time.Sleep(time.Second)
				callCnt = 0
			}
		}

		totalCnt += 1
		lr.lastRefreshSec = nowSec
	}
}

func (lr *LiquidityRefresher) watchPostProcess() {
	pendingCount := 0
	for {
		tag := <-lr.postProcess
		pendingCount += 1
		if pendingCount%100 == 0 {
			log.Printf("Processed %d total liq refreshes. Last=%s. len(urgent)=%d, len(slow)=%d", pendingCount, tag, len(lr.workUrgent), len(lr.workSlow))
		}
	}
}
