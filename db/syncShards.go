package db

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"cloud.google.com/go/storage"
	"github.com/CrocSwap/graphcache-go/loader"
)

var bucketName = "gcgo-swap-shards"

func SyncLocalShardsWithUniswap(chainCfg loader.ChainConfig) {
	chainCfg.Subgraph = "https://api.thegraph.com/subgraphs/name/uniswap/uniswap-v3"

	// Example usage
	initialTimestamp := 1672531200 // January 1, 2023

	daysList := GetDaysList(int64(initialTimestamp))	
	shards, err := FetchBucketItems(bucketName)
	if err != nil {
		log.Println("[Shard Syncer]: Error ", err)
		return
	}
	if err != nil {
		log.Fatalf("Failed to get files in directory: %v", err)
	}
	for _, day := range daysList {
		startTime := GetStartOfDayTimestamp(day)
		endTime := GetEndOfDayTimestamp(day)
		shardPath := fmt.Sprintf("./db/shards/%s", day)
		fullShardPath := shardPath + ".db"
		if(fileExistsInDir(fullShardPath)){
			if(!filePathExistsInBucket(fullShardPath, shards)){
				log.Println("[Shard Syncer]: Shard exists in locally but not in GCS, uploading ", shardPath)
				UploadShardToBucket(fullShardPath, shards)
			} else {
				log.Println("[Shard Syncer]: Shard exists locally, skipping ", shardPath)
			}
		}else if(filePathExistsInBucket(fullShardPath, shards)){
			log.Println("[Shard Syncer]: Shard exists in GCS, downloading ", shardPath)
			DownloadShardFromBucket(fullShardPath)
		}else{
			log.Println("[Shard Syncer]: Creating shard from uniswap ", shardPath)
			FetchUniswapAndSaveToShard(chainCfg,shardPath, int(startTime), int(endTime))
		}

		
	}
}

func DownloadAllShardsInBucket(){
	shards, err := FetchBucketItems(bucketName)
	if err != nil {
		log.Println("[Shard Syncer]: Error ", err)
		return 
	}
	log.Println("[Shard Syncer]: Found ", len(shards), " shards")
}

func DownloadShardFromBucket(filePath string){
	_, fileName := filepath.Split(filePath)
	data, err := FetchObjectData(bucketName, fileName)
	if err != nil {
		log.Fatalf("Failed to fetch object data: %v", err)
	}
	// Now write the data to the filepath
	err = ioutil.WriteFile(filePath, data, 0644)
	if err != nil {
		log.Fatalf("Failed to write file: %v", err)
	}

	log.Println("[Shard Syncer]: Finished - Downloaded shard from bucket", filePath)


}

func UploadAllShardsToBucket(){
	directoryPath := "./db/shards" // Specify the path to the directory
	shards, err := FetchBucketItems(bucketName)
	if err != nil {
		log.Fatalf("Failed to get shards from GCS: %v", err)
	}
	files, err := getFilesInDirectory(directoryPath)
	if err != nil {
		log.Fatalf("Failed to get files in directory: %v", err)
	}

	for _, file := range files {
		filePath := directoryPath + "/" + file.Name()
		objectName := file.Name()
		// If file isn't in list of shards, upload it
		if fileExistsInBucket(objectName, shards) == false {
			log.Printf("[Shard Syncer]: In Progress - Uploading '%s' to bucket '%s'\n", filePath, bucketName)
			err := UploadItemToBucket(bucketName, objectName, filePath)
			if err != nil {
				log.Printf("[Shard Syncer]: Failed to upload '%s': %v\n", filePath, err)
			} else {
				log.Printf("[Shard Syncer]: Finished - Uploaded '%s' to bucket '%s'\n", filePath, bucketName)
			}
		} else {
			log.Printf("[Shard Syncer]: Skipping '%s' to bucket '%s'\n", filePath, bucketName)
		}
	}

	log.Println("[Shard Syncer]: Finished - Uploaded all shards to bucket")
}


func UploadShardToBucket(filePath string, shards []*storage.ObjectAttrs){
	_, fileName := filepath.Split(filePath)
	var err error
	if(len(shards) == 0){
		
		shards, err = FetchBucketItems(bucketName)
		if err != nil {
			log.Fatalf("Failed to get shards from GCS: %v", err)
		}
	}
	
	// If file isn't in list of shards, upload it
	if fileExistsInBucket(fileName, shards) == false {
		log.Printf("[Shard Syncer]: In Progress - Uploading '%s' to bucket '%s'\n", filePath, bucketName)
		err := UploadItemToBucket(bucketName, fileName, filePath)
		if err != nil {
			log.Printf("[Shard Syncer]: Failed to upload '%s': %v\n", filePath, err)
		} else {
			log.Printf("[Shard Syncer]: Finished - Uploaded '%s' to bucket '%s'\n", filePath, bucketName)
		}
	} else {
		log.Printf("[Shard Syncer]: Skipping '%s' to bucket '%s'\n", filePath, bucketName)
	}
	
}

func filePathExistsInBucket(filePath string, shards []*storage.ObjectAttrs) bool {
	_, fileName := filepath.Split(filePath)
	return fileExistsInBucket(fileName, shards)
}
func fileExistsInBucket(fileName string, shards []*storage.ObjectAttrs) bool {
	for _, shard := range shards {
		objectName := shard.Name
		if objectName == fileName {
			return true
		}
	}

	return false
}

func fileExistsInDir(filePath string) bool {
	// Check if the file exists
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false
	}

	return true
}
func getFilesInDirectory(directoryPath string) ([]os.FileInfo, error) {
	files, err := ioutil.ReadDir(directoryPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %v", err)
	}

	var fileList []os.FileInfo
	for _, file := range files {
		if !file.IsDir() {
			fileList = append(fileList, file)
		}
	}

	return fileList, nil
}

// GetDaysList returns a list of formatted date strings between the given initial timestamp and the current date (excluding the current day).
func GetDaysList(initialTimestamp int64) []string {
	currentTime := time.Now()
	currentDate := currentTime.Truncate(24 * time.Hour) // Truncate time to get the start of the current day

	var daysList []string

	for timestamp := initialTimestamp + 24*60*60; timestamp < currentDate.Unix(); timestamp += 24*60*60 {
		date := time.Unix(timestamp, 0).Format("2006-01-02")
		daysList = append(daysList, date)
	}

	return daysList
}

// GetStartOfDayTimestamp returns the Unix timestamp representing the start of the given date.
func GetStartOfDayTimestamp(date string) int64 {
	startTime, _ := time.Parse("2006-01-02", date)
	return startTime.Unix()
}

// GetEndOfDayTimestamp returns the Unix timestamp representing the end of the given date.
func GetEndOfDayTimestamp(date string) int64 {
	endTime, _ := time.Parse("2006-01-02", date)
	return endTime.Add(24*time.Hour - time.Second).Unix()
}

// FetchSwapShard is a placeholder function that accepts start and end timestamps and performs some operation.
func FetchSwapShard(shardName string, startTime, endTime int) {
	fmt.Printf("Calling FetchSwapShard for %s with start time: %v, end time: %v\n", shardName, startTime, endTime)
}
