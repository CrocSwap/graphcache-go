#!/bin/bash

# Load the .env file
# Should contain these vars:
# DOCKERHUB_USERNAME=your-dockerhub-username
# DOCKERHUB_PASSWORD=your-dockerhub-password
source .env

# Variables
IMAGE_NAME="graphcache-go"
IMAGE_TAG="latest"


# print DOCKERHUB_USERNAME
echo "DOCKERHUB_USERNAME: $DOCKERHUB_USERNAME"
echo "DOCKERHUB_PASSWORD: $DOCKERHUB_PASSWORD"
# Navigate to the directory where your Dockerfile is located
cd /path/to/your/dockerfile/directory

# Build the Docker image
echo "Building Docker image..."
docker build --platform linux/amd64 -t $DOCKERHUB_USERNAME/$IMAGE_NAME:$IMAGE_TAG .


# Login to DockerHub
echo "Logging in to DockerHub..."
echo $DOCKERHUB_PASSWORD | docker login -u $DOCKERHUB_USERNAME --password-stdin

# Push the Docker image
echo "Pushing Docker image to DockerHub..."
docker push $DOCKERHUB_USERNAME/$IMAGE_NAME:$IMAGE_TAG

# Logout from DockerHub
echo "Logging out from DockerHub..."
docker logout

echo "Done!"
