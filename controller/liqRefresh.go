package controller

import (
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
	posQueries chan *PositionRefresher
	query      *loader.CrocQuery
}

const QUERY_CHANNEL_WINDOW = 25000
const POSITION_CHANNEL_WINDOW = 1000

func NewLiquidityRefresher(query *loader.CrocQuery) *LiquidityRefresher {
	refresher := LiquidityRefresher{
		posQueries: make(chan *PositionRefresher, QUERY_CHANNEL_WINDOW),
		query:      query,
	}
	go refresher.watchPending()
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
		posType := types.PositionTypeForLiq(posRefresher.location.LiquidityLocation)

		if posType == "ambient" {
			ambientSeeds, err := r.query.QueryAmbientSeeds(posRefresher.location)
			if err == nil {
				defer posRefresher.lock.Unlock()
				posRefresher.lock.Lock()
				posRefresher.pos.UpdateAmbient(*ambientSeeds)
			}
		}

		if posType == "range" {
			concLiq, err := r.query.QueryRangeLiquidity(posRefresher.location)
			rewardLiq, err2 := r.query.QueryRangeRewardsLiq(posRefresher.location)

			if err == nil && err2 == nil {
				defer posRefresher.lock.Unlock()
				posRefresher.lock.Lock()
				posRefresher.pos.UpdateRange(*concLiq, *rewardLiq)
			}
		}

		if posType == "knockout" {
			concLiq, isKnockedOut, err := r.query.QueryKnockoutLiq(posRefresher.location)
			if err == nil {
				defer posRefresher.lock.Unlock()
				posRefresher.lock.Lock()
				posRefresher.pos.UpdateKnockout(*concLiq, isKnockedOut)
			}
		}
	}
}
