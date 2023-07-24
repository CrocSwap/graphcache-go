package db

import (
	"context"
	"io/ioutil"
	"log"
	"os"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)
var credentialsFile = "./db/GCS_credentials.json"

// FetchBucketItems fetches the items (objects) from the specified GCS bucket and returns a slice of storage.ObjectAttrs.
func FetchBucketItems(bucketName string) ([]*storage.ObjectAttrs, error) {
	log.Println("[Shard Syncer]: Fetching shards from GCS")
	var ctx = context.Background()
	var client, err = storage.NewClient(ctx, option.WithCredentialsFile(credentialsFile))
	
	if err != nil {
		return nil, err
	}

	it := client.Bucket(bucketName).Objects(ctx, nil)
	var items []*storage.ObjectAttrs

	for {
		obj, err := it.Next()
		if err != nil {
			break
		}
		if err != nil {
			return nil, err
		}

		items = append(items, obj)
	}

	return items, nil
}

// UploadItemToBucket uploads a local file to the specified GCS bucket with the given object name.
func UploadItemToBucket(bucketName, objectName, filePath string) error {
	var ctx = context.Background()
	var client, err = storage.NewClient(ctx, option.WithCredentialsFile(credentialsFile))
	
	if err != nil {
		return err
	}

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	wc := client.Bucket(bucketName).Object(objectName).NewWriter(ctx)
	if _, err := wc.Write(readFile(file)); err != nil {
		return err
	}
	if err := wc.Close(); err != nil {
		return err
	}

	return nil
}

func readFile(file *os.File) []byte {
	content, _ := ioutil.ReadAll(file)
	return content
}

// FetchObjectData fetches the data of a single object from the specified GCS bucket.
func FetchObjectData(bucketName, objectName string) ([]byte, error) {
	log.Println("[Shard Syncer]: Fetching shard from GCS", objectName)
	var ctx = context.Background()
	var client, err = storage.NewClient(ctx, option.WithCredentialsFile(credentialsFile))
	
	if err != nil {
		return nil, err
	}

	reader, err := client.Bucket(bucketName).Object(objectName).NewReader(ctx)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return data, nil
}


