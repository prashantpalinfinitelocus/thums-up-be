#!/bin/bash

set -e

PROJECT_ID=${1:-""}
if [ -z "$PROJECT_ID" ]; then
    echo "Usage: ./setup-gcp.sh PROJECT_ID"
    exit 1
fi

echo "Setting up GCP infrastructure for project: $PROJECT_ID"

echo "Enabling required APIs..."
gcloud services enable \
    cloudrun.googleapis.com \
    cloudbuild.googleapis.com \
    pubsub.googleapis.com \
    storage-api.googleapis.com \
    sqladmin.googleapis.com \
    secretmanager.googleapis.com \
    artifactregistry.googleapis.com \
    --project=$PROJECT_ID

echo "APIs enabled successfully"

echo "Creating Terraform backend bucket..."
gsutil mb -p $PROJECT_ID -l asia-south1 gs://${PROJECT_ID}-terraform-state || echo "Bucket already exists"
gsutil versioning set on gs://${PROJECT_ID}-terraform-state

echo "Initializing Terraform..."
cd terraform
terraform init

echo "Planning Terraform deployment..."
terraform plan -var="project_id=$PROJECT_ID"

echo ""
echo "Review the plan above. To apply, run:"
echo "cd terraform && terraform apply -var=\"project_id=$PROJECT_ID\""

