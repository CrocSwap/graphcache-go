package db

import (
	"log"
	"time"

	"github.com/CrocSwap/graphcache-go/loader"
)
func ScheduleSyncShards(hourToSyncUniswapShards int, chainCfg loader.ChainConfig) {
	if(hourToSyncUniswapShards >= 24){
		log.Println("[Shard Syncer]: Invalid hour to sync uniswap shards, defaulting to 23")
		hourToSyncUniswapShards = 23
	}
	if(hourToSyncUniswapShards < 0){
		log.Println("[Shard Syncer]: Invalid hour to sync uniswap shards, defaulting to 0")
		hourToSyncUniswapShards = 0
	}
	for {
		// Calculate the duration until the specified time after 1 AM
		now := time.Now()
		year, month, day := now.Date()
		nextTime := time.Date(year, month, day, hourToSyncUniswapShards, 0, 0, 0, time.Local)
		if nextTime.Before(now) {
			nextTime = nextTime.Add(24 * time.Hour) // If the specified time has already passed today, schedule it for the next day
		}
		durationUntilNextTime := nextTime.Sub(now)
		log.Println("[Shard Syncer]: Next sync scheduled: ", durationUntilNextTime)

		// Wait for the calculated duration
		timer := time.NewTimer(durationUntilNextTime)
		<-timer.C

		// Execute the task
		log.Println("[Shard Syncer]: Executing shard syncer")
		go SyncLocalShardsWithUniswap(chainCfg)

		// Reset the timer to schedule the next task 
		timer.Reset(24 * time.Second)
		
	}
}
