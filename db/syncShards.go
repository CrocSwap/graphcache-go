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



func SyncLocalShardsWithUniswap(chainCfg loader.ChainConfig) {
	daysList := GetDaysList(InitialTimestamp)
	shards, err := FetchBucketItems(BucketName)
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
		shardPath := fmt.Sprintf("%s/%s",ShardsPath, day)
		fullShardPath := shardPath + ".db"
		if(FileExistsInDir(fullShardPath)){
			if(!FilePathExistsInBucket(fullShardPath, shards)){
				log.Println("[Shard Syncer]: Shard exists in locally but not in GCS, uploading ", shardPath)
				UploadShardToBucket(fullShardPath)
			} else {
				log.Println("[Shard Syncer]: Shard exists locally, skipping ", shardPath)
			}
		}else if(FilePathExistsInBucket(fullShardPath, shards)){
			log.Println("[Shard Syncer]: Shard exists in GCS, downloading ", shardPath)
			DownloadShardFromBucket(fullShardPath)
		}else{
			log.Println("[Shard Syncer]: Creating shard from uniswap ", shardPath)
			FetchUniswapAndSaveToShard(chainCfg,shardPath, int(startTime), int(endTime), "Shard Syncer")
		}

		
	}
}


func DownloadShardFromBucket(filePath string){
	_, fileName := filepath.Split(filePath)
	data, err := FetchObjectData(BucketName, fileName)
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
	shards, err := FetchBucketItems(BucketName)
	if err != nil {
		log.Fatalf("Failed to get shards from GCS: %v", err)
	}
	files, err := GetFilesInDirectory(directoryPath)
	if err != nil {
		log.Fatalf("Failed to get files in directory: %v", err)
	}

	for _, file := range files {
		filePath := directoryPath + "/" + file.Name()
		objectName := file.Name()
		// If file isn't in list of shards, upload it
		if fileExistsInBucket(objectName, shards) == false {
			log.Printf("[Shard Syncer]: In Progress - Uploading '%s' to bucket '%s'\n", filePath, BucketName)
			err := UploadItemToBucket(BucketName, objectName, filePath)
			if err != nil {
				log.Printf("[Shard Syncer]: Failed to upload '%s': %v\n", filePath, err)
			} else {
				log.Printf("[Shard Syncer]: Finished - Uploaded '%s' to bucket '%s'\n", filePath, BucketName)
			}
		} else {
			log.Printf("[Shard Syncer]: Skipping '%s' to bucket '%s'\n", filePath, BucketName)
		}
	}

	log.Println("[Shard Syncer]: Finished - Uploaded all shards to bucket")
}


func UploadShardToBucket(filePath string){
	_, fileName := filepath.Split(filePath)
		
	shards, err := FetchBucketItems(BucketName)
	if err != nil {
		log.Fatalf("Failed to get shards from GCS: %v", err)
	}
	
	// If file isn't in list of shards, upload it
	if fileExistsInBucket(fileName, shards) == false {
		log.Printf("[Shard Syncer]: In Progress - Uploading '%s' to bucket '%s'\n", filePath, BucketName)
		err := UploadItemToBucket(BucketName, fileName, filePath)
		if err != nil {
			log.Printf("[Shard Syncer]: Failed to upload '%s': %v\n", filePath, err)
		} else {
			log.Printf("[Shard Syncer]: Finished - Uploaded '%s' to bucket '%s'\n", filePath, BucketName)
		}
	} else {
		log.Printf("[Shard Syncer]: Skipping '%s' to bucket '%s'\n", filePath, BucketName)
	}
	
}

func FilePathExistsInBucket(filePath string, shards []*storage.ObjectAttrs) bool {
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

func FileExistsInDir(filePath string) bool {
	// Check if the file exists
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false
	}

	return true
}
func GetFilesInDirectory(directoryPath string) ([]os.FileInfo, error) {
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
func GetDaysList(startTime int64) []string {
	currentTime := time.Now()
	currentDate := currentTime.Truncate(24 * time.Hour) // Truncate time to get the start of the current day

	var daysList []string

	for timestamp := currentDate.Unix() - 1; timestamp >= startTime;  timestamp -= 24*60*60 {
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
