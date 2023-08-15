#!/bin/bash

# Step 1: Pull the Docker image
IMAGE="cadehypotenuse/graphcache-go-candles:latest"
echo "Pulling image $IMAGE..."
docker pull $IMAGE

# Step 2: Create the 'shards' folder in root if it doesn't exist
if [ ! -d "/shards" ]; then
  echo "Creating 'shards' directory..."
  mkdir /shards
else
  echo "'shards' directory already exists, skipping creation."
fi

# Step 3: Check if 'GCS_credentials.json' exists, log error if not
if [ ! -f "GCS_credentials.json" ]; then
  echo "Error: GCS_credentials.json does not exist."
  exit 1
else
  echo "GCS_credentials.json found."
fi

# Step 4: Run the image as a container
echo "Running the container..."
docker-compose -f docker-compose.uniswap.prod.yml up

echo "Script execution completed."