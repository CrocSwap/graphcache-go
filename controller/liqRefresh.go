package controller

import (
	"log"
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

const NUM_PARALLEL_QUERIES = 250
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

func (r *LiquidityRefresher) watchPending() {
	for true {
		posRefresher := <-r.posQueries
		r.workers[r.nextWorker] <- posRefresher

		r.nextWorker = r.nextWorker + 1
		if r.nextWorker == len(r.workers) {
			r.nextWorker = 0
		}
	}
}

const RETRY_QUERY_DURATION = 10 * time.Second
const N_MAX_RETRIES = 3

func (r *LiquidityRefresher) watchWorker(workQueue chan *PositionRefresher, postProcess chan *PositionRefresher) {
	for true {
		posRefresher := <-workQueue
		posType := types.PositionTypeForLiq(posRefresher.location.LiquidityLocation)

		if posType == "ambient" {
			ambientSeeds, err := r.query.QueryAmbientSeeds(posRefresher.location)

			for retryCount := 0; err != nil && retryCount < N_MAX_RETRIES; retryCount += 1 {
				time.Sleep(RETRY_QUERY_DURATION)
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

			for retryCount := 0; err != nil && retryCount < N_MAX_RETRIES; retryCount += 1 {
				time.Sleep(RETRY_QUERY_DURATION)
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
