package db

import (
	"encoding/json"
	"log"

	"github.com/CrocSwap/graphcache-go/utils"
)

var BucketName = utils.GoDotEnvVariable("UNISWAP_GCS_BUCKET_NAME"); 
var ShardsPath = utils.GoDotEnvVariable("UNISWAP_SHARDS_PATH"); 
var	InitialTimestamp = int64(utils.GetEnvVarIntFromString("UNISWAP_INITIAL_TIMESTAMP", 1672531200))
var DaysOfCandlesBeforeServerReady = utils.GetEnvVarIntFromString("UNISWAP_DAYS_OF_CANDLES_BEFORE_SERVER_READY", 0)




var GCSCredentials = map[string]string{
	"type":                        utils.GoDotEnvVariable("GCS_TYPE"),
	"project_id":                  utils.GoDotEnvVariable("GCS_PROJECT_ID"),
	"private_key_id":              utils.GoDotEnvVariable("GCS_PRIVATE_KEY_ID"),
	"private_key":                 utils.GoDotEnvVariable("GCS_PRIVATE_KEY"),
	"client_email":                utils.GoDotEnvVariable("GCS_CLIENT_EMAIL"),
	"client_id":                   utils.GoDotEnvVariable("GCS_CLIENT_ID"),
	"auth_uri":                    utils.GoDotEnvVariable("GCS_AUTH_URI"),
	"token_uri":                   utils.GoDotEnvVariable("GCS_TOKEN_URI"),
	"auth_provider_x509_cert_url": utils.GoDotEnvVariable("GCS_AUTH_PROVIDER_X509_CERT_URL"),
	"client_x509_cert_url":        utils.GoDotEnvVariable("GCS_CLIENT_X509_CERT_URL"),
	"universe_domain":             utils.GoDotEnvVariable("GCS_UNIVERSE_DOMAIN"),
}

func GetGCSCredentials() []byte{
	creds, err := json.Marshal(GCSCredentials)
	if err != nil {
		log.Println("[Shard Syncer]: Error marshaling credentials", err)
		return nil
	}
	
	return creds
}

