#!/bin/bash

if [[ $# -ge 1 ]] ; then
    dbPath=$1
else
    mkdir -p ./_data/
fi

BUCKET_NAME="hypelabs-public/crocswap"
FILE_NAME="database.db"
SAVE_PATH="_data/"

# Create the directory if it doesn't exist
mkdir -p $SAVE_PATH

# Download the file from the S3 bucket to the specified folder
aws s3 cp s3://$BUCKET_NAME/$FILE_NAME $SAVE_PATH

# Check if the download was successful
if [ $? -eq 0 ]; then
    echo "File downloaded successfully."
else
    echo "Error downloading file."
fi

