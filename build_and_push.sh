#!/bin/bash

# Exit immediately if any command fails
set -e

# Default version tag
DEFAULT_TAG="v0.0.6.5"
TAG=${1:-$DEFAULT_TAG}

# Configurable AWS variables
AWS_REGION="ap-south-1"
AWS_ACCOUNT_ID="591051854019"
ECR_REGISTRY="${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com"
IMAGE_NAME="shiv-mitra"
FULL_TAG="${TAG}-attendance_cmrf-uat"

echo "=========================================================="
echo "Starting Docker Build & Push Workflow"
echo "Image: ${IMAGE_NAME}"
echo "Tag: ${FULL_TAG}"
echo "ECR Registry: ${ECR_REGISTRY}"
echo "=========================================================="

# Step 1: AWS ECR Authentication
echo "🔑 Logging in to AWS ECR..."
aws ecr get-login-password --region ${AWS_REGION} | docker login --username AWS --password-stdin ${ECR_REGISTRY}

# Step 2: Build the Docker Image
echo "🔨 Building Docker image: ${IMAGE_NAME}:${FULL_TAG}..."
docker build -t ${IMAGE_NAME}:${FULL_TAG} .

# Step 3: Tag the Image for ECR
echo "🏷️ Tagging image for ECR registry..."
docker tag ${IMAGE_NAME}:${FULL_TAG} ${ECR_REGISTRY}/${IMAGE_NAME}:${FULL_TAG}

# Step 4: Push to AWS ECR
echo "🚀 Pushing image to ECR..."
docker push ${ECR_REGISTRY}/${IMAGE_NAME}:${FULL_TAG}

echo "✅ Success! Image pushed to ECR: ${ECR_REGISTRY}/${IMAGE_NAME}:${FULL_TAG}"
