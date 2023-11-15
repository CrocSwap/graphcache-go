package controller

import (
	"log"
	"math/rand"
	"time"

	"github.com/CrocSwap/graphcache-go/loader"
)

type HandleRefresher struct {
	hndl         IRefreshHandle
	requests     chan int64
	pendingQueue chan IRefreshHandle
}

type LiquidityRefresher struct {
	pending     chan IRefreshHandle
	postProcess chan string
	query       *loader.ICrocQuery
	workers     []chan IRefreshHandle
	nextWorker  int
}

const NUM_PARALLEL_QUERIES = 50
const QUERY_CHANNEL_WINDOW = 25000
const POSITION_CHANNEL_WINDOW = 1000
const QUERY_WORKER_QUEUE = 1000

func NewLiquidityRefresher(query *loader.ICrocQuery) *LiquidityRefresher {
	workers := make([]chan IRefreshHandle, NUM_PARALLEL_QUERIES)
	for idx := 0; idx < NUM_PARALLEL_QUERIES; idx += 1 {
		workers[idx] = make(chan IRefreshHandle, QUERY_WORKER_QUEUE)
	}

	refresher := LiquidityRefresher{
		pending:     make(chan IRefreshHandle, QUERY_CHANNEL_WINDOW),
		query:       query,
		workers:     workers,
		postProcess: make(chan string, QUERY_CHANNEL_WINDOW),
		nextWorker:  0,
	}

	go refresher.watchPending()
	go refresher.watchPostProcess()
	for _, worker := range refresher.workers {
		go refresher.watchWork(worker)
	}

	return &refresher
}

func NewHandleRefresher(hndl IRefreshHandle, queue chan IRefreshHandle) *HandleRefresher {
	refresher := &HandleRefresher{
		requests:     make(chan int64, POSITION_CHANNEL_WINDOW),
		hndl:         hndl,
		pendingQueue: queue,
	}
	go refresher.watchPending()
	return refresher
}

const REFRESH_WINDOW = 15

func (r *HandleRefresher) PushRefresh(eventTime int) {
	r.requestRefresh()
	if isRecentEvent(eventTime) {
		go r.pushFollowup()
	}
}

func (r *HandleRefresher) requestRefresh() {
	latestTime := time.Now().Unix()
	windowTag := latestTime / REFRESH_WINDOW
	r.requests <- windowTag
}

/* Historical events won't require followup, because an RPC call should
 * be sync'd */
const RECENT_EVENT_WINDOW = 60

func isRecentEvent(eventTime int) bool {
	currentTime := time.Now().Unix()
	return int(currentTime)-eventTime < RECENT_EVENT_WINDOW
}

/* Followup after refresh, in case RPC node hasn't synced to last
 * block at first call. */
func (r *HandleRefresher) pushFollowup() {
	REFRESH_FOLLOWUP_SECS := []time.Duration{2, 10, 30, 60}
	for _, interval := range REFRESH_FOLLOWUP_SECS {
		time.Sleep(interval * time.Second)
		r.requestRefresh()
	}
}

func (r *HandleRefresher) watchPending() {
	prevLatest := int64(0)

	for true {
		latestWindow := <-r.requests
		if latestWindow > prevLatest {
			r.pendingQueue <- r.hndl
		}
		prevLatest = latestWindow
	}
}

const MAX_REQS_PER_SEC = 50

func (r *LiquidityRefresher) watchPending() {
	lastSec := time.Now().Unix()
	callCnt := 0

	for true {
		nowSec := time.Now().Unix()
		if nowSec == lastSec {
			callCnt += 1
		} else {
			callCnt = 0
			nowSec = lastSec
		}

		if callCnt > MAX_REQS_PER_SEC {
			log.Println("Throttling liquidity refreshes per second")
			time.Sleep(time.Second)

		} else {
			posRefresher := <-r.pending
			r.workers[r.nextWorker] <- posRefresher

			r.nextWorker = r.nextWorker + 1
			if r.nextWorker == len(r.workers) {
				r.nextWorker = 0
			}
		}
	}
}

const RETRY_QUERY_MIN_WAIT = 30
const RETRY_QUERY_MAX_WAIT = 60
const N_MAX_RETRIES = 3

// Do this so that in case the problem is overloading the RPC, calls don't all spam again
// at same deterministic time
func retryWaitRandom() {
	waitTime := rand.Intn(RETRY_QUERY_MAX_WAIT-RETRY_QUERY_MIN_WAIT) + RETRY_QUERY_MIN_WAIT
	log.Printf("Query attempt failed. Retrying again in %d seconds", waitTime)
	time.Sleep(time.Duration(waitTime) * time.Second)
}

func (r *LiquidityRefresher) watchWork(workQueue chan IRefreshHandle) {
	for true {
		hndl := <-workQueue
		hndl.RefreshQuery(r.query)
		r.postProcess <- hndl.LabelTag()
	}
}

func (r *LiquidityRefresher) watchPostProcess() {
	pendingCount := 0
	for true {
		tag := <-r.postProcess
		pendingCount += 1
		if pendingCount%100 == 0 {
			log.Printf("Processed %d liquidity refreshes since startup. Last type=%s", pendingCount, tag)
		}
	}
}
