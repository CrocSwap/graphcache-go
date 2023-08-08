package db

import (
	"strconv"

	"github.com/CrocSwap/graphcache-go/utils"
)

var BucketName = utils.GoDotEnvVariable("UNISWAP_GCS_BUCKET_NAME"); 
var ShardsPath = utils.GoDotEnvVariable("UNISWAP_SHARDS_PATH"); 
var	InitialTimestamp = int64(getEnvVarIntFromString("UNISWAP_INITIAL_TIMESTAMP", 1672531200))
var DaysOfCandlesBeforeServerReady = getEnvVarIntFromString("UNISWAP_DAYS_OF_CANDLES_BEFORE_SERVER_READY", 0)
var credentialsFile = utils.GoDotEnvVariable("UNISWAP_PATH_TO_GCS_CREDENTIALS"); 


func getEnvVarIntFromString(envVar string, fallback int) int {
	var envVarInt, err = strconv.Atoi(utils.GoDotEnvVariable(envVar))
	if(err != nil){
		return fallback
	}
	return envVarInt
}
