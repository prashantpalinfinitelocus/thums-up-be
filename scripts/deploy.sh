#!/bin/bash

set -e

SERVICE=${1:-"main"}
PROJECT_ID=${2:-""}

if [ -z "$PROJECT_ID" ]; then
    echo "Usage: ./deploy.sh [main|subscriber] PROJECT_ID"
    exit 1
fi

if [ "$SERVICE" == "main" ]; then
    CONFIG_FILE="cloudbuild-main.yaml"
    echo "Deploying main backend service..."
elif [ "$SERVICE" == "subscriber" ]; then
    CONFIG_FILE="cloudbuild-subscriber.yaml"
    echo "Deploying subscriber service..."
else
    echo "Invalid service. Use 'main' or 'subscriber'"
    exit 1
fi

gcloud builds submit \
    --config=$CONFIG_FILE \
    --project=$PROJECT_ID \
    --substitutions=_PROJECT_ID=$PROJECT_ID

echo "$SERVICE service deployed successfully!"

