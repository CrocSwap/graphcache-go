package controller

import (
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/CrocSwap/graphcache-go/loader"
	"github.com/CrocSwap/graphcache-go/model"
	"github.com/CrocSwap/graphcache-go/types"
)

type PositionRefresher struct {
	location types.PositionLocation
	requests chan int64
	queries  chan *PositionRefresher
	lock     sync.RWMutex
	pos      *model.PositionTracker
}

type LiquidityRefresher struct {
	posQueries  chan *PositionRefresher
	postProcess chan *PositionRefresher
	query       *loader.CrocQuery
	workers     []chan *PositionRefresher
	nextWorker  int
}

const NUM_PARALLEL_QUERIES = 100
const QUERY_CHANNEL_WINDOW = 25000
const POSITION_CHANNEL_WINDOW = 1000
const QUERY_WORKER_QUEUE = 1000

func NewLiquidityRefresher(query *loader.CrocQuery) *LiquidityRefresher {
	workers := make([]chan *PositionRefresher, NUM_PARALLEL_QUERIES)

	for idx := 0; idx < NUM_PARALLEL_QUERIES; idx += 1 {
		workers[idx] = make(chan *PositionRefresher, QUERY_WORKER_QUEUE)
	}

	refresher := LiquidityRefresher{
		posQueries:  make(chan *PositionRefresher, QUERY_CHANNEL_WINDOW),
		query:       query,
		workers:     workers,
		postProcess: make(chan *PositionRefresher, QUERY_CHANNEL_WINDOW),
		nextWorker:  0,
	}

	go refresher.watchPending()
	go refresher.watchPostProcess()
	for _, worker := range refresher.workers {
		go refresher.watchWorker(worker, refresher.postProcess)
	}

	return &refresher
}

func NewPositionRefresher(loc types.PositionLocation, liq *LiquidityRefresher,
	pos *model.PositionTracker) *PositionRefresher {
	refresher := &PositionRefresher{
		location: loc,
		requests: make(chan int64, POSITION_CHANNEL_WINDOW),
		pos:      pos,
		queries:  liq.posQueries,
	}
	go refresher.watchPending()
	return refresher
}

const REFRESH_WINDOW = 15

func (r *PositionRefresher) PushRefresh() {
	latestTime := time.Now().Unix()
	windowTag := latestTime / REFRESH_WINDOW
	r.requests <- windowTag
}

func (r *PositionRefresher) watchPending() {
	prevLatest := int64(0)

	for true {
		latestWindow := <-r.requests
		if latestWindow > prevLatest {
			r.queries <- r
		}
		prevLatest = latestWindow
	}
}

const MAX_REQS_PER_SEC = 200

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
			posRefresher := <-r.posQueries
			r.workers[r.nextWorker] <- posRefresher

			r.nextWorker = r.nextWorker + 1
			if r.nextWorker == len(r.workers) {
				r.nextWorker = 0
			}
		}
	}
}

const RETRY_QUERY_MIN_WAIT = 5
const RETRY_QUERY_MAX_WAIT = 15
const N_MAX_RETRIES = 3

// Do this so that in case the problem is overloading the RPC, calls don't all spam again
// at same deterministic time
func retryWaitRandom() {
	waitTime := rand.Intn(RETRY_QUERY_MAX_WAIT-RETRY_QUERY_MIN_WAIT) + RETRY_QUERY_MIN_WAIT
	log.Printf("Query attempt failed. Retrying again in %d seconds", waitTime)
	time.Sleep(time.Duration(waitTime) * time.Second)
}

func (r *LiquidityRefresher) watchWorker(workQueue chan *PositionRefresher, postProcess chan *PositionRefresher) {
	for true {
		posRefresher := <-workQueue
		posType := types.PositionTypeForLiq(posRefresher.location.LiquidityLocation)

		if posType == "ambient" {
			ambientSeeds, err := r.query.QueryAmbientSeeds(posRefresher.location)

			for retryCount := 0; err != nil && retryCount < N_MAX_RETRIES; retryCount += 1 {
				retryWaitRandom()
				ambientSeeds, err = r.query.QueryAmbientSeeds(posRefresher.location)
			}
			if err != nil {
				log.Fatal("Unable to sync liquidity on ambient order")
			}

			defer posRefresher.lock.Unlock()
			posRefresher.lock.Lock()
			posRefresher.pos.UpdateAmbient(*ambientSeeds)
		}

		if posType == "range" {
			concLiq, err := r.query.QueryRangeLiquidity(posRefresher.location)
			rewardLiq, err2 := r.query.QueryRangeRewardsLiq(posRefresher.location)

			for retryCount := 0; (err != nil || err2 != nil) && retryCount < N_MAX_RETRIES; retryCount += 1 {
				retryWaitRandom()
				concLiq, err = r.query.QueryAmbientSeeds(posRefresher.location)
				rewardLiq, err2 = r.query.QueryRangeRewardsLiq(posRefresher.location)
			}
			if err != nil || err2 != nil {
				log.Fatal("Unable to sync liquidity on range order")
			}

			defer posRefresher.lock.Unlock()
			posRefresher.lock.Lock()
			posRefresher.pos.UpdateRange(*concLiq, *rewardLiq)
		}

		if posType == "knockout" {
			/* concLiq, isKnockedOut, err := r.query.QueryKnockoutLiq(posRefresher.location)
			if err == nil {
				defer posRefresher.lock.Unlock()
				posRefresher.lock.Lock()
				posRefresher.pos.UpdateKnockout(*concLiq, isKnockedOut)
			}*/
		}

		postProcess <- posRefresher
	}
}

func (r *LiquidityRefresher) watchPostProcess() {
	pendingCount := 0
	for true {
		<-r.postProcess
		pendingCount += 1
		if pendingCount%100 == 0 {
			log.Printf("Processed %d liquidity refreshes since startup", pendingCount)
		}
	}
}
