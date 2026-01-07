terraform {
  required_version = ">= 1.0"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }

  backend "gcs" {
    bucket = "thums-up-terraform-state"
    prefix = "terraform/state"
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}

resource "google_project_service" "services" {
  for_each = toset([
    "cloudrun.googleapis.com",
    "cloudbuild.googleapis.com",
    "pubsub.googleapis.com",
    "storage-api.googleapis.com",
    "firestore.googleapis.com",
    "sqladmin.googleapis.com",
    "secretmanager.googleapis.com",
  ])

  service            = each.key
  disable_on_destroy = false
}

resource "google_sql_database_instance" "postgres" {
  name             = "${var.project_name}-postgres"
  database_version = "POSTGRES_15"
  region           = var.region

  settings {
    tier = "db-f1-micro"

    ip_configuration {
      ipv4_enabled = true
      authorized_networks {
        name  = "all"
        value = "0.0.0.0/0"
      }
    }

    backup_configuration {
      enabled = true
    }
  }

  deletion_protection = false
}

resource "google_sql_database" "database" {
  name     = var.database_name
  instance = google_sql_database_instance.postgres.name
}

resource "google_sql_user" "user" {
  name     = var.database_user
  instance = google_sql_database_instance.postgres.name
  password = var.database_password
}

resource "google_storage_bucket" "bucket" {
  name          = "${var.project_name}-storage"
  location      = var.region
  force_destroy = false

  uniform_bucket_level_access = true

  cors {
    origin          = ["*"]
    method          = ["GET", "POST", "PUT", "DELETE"]
    response_header = ["*"]
    max_age_seconds = 3600
  }
}

resource "google_pubsub_topic" "main_topic" {
  name = "${var.project_name}-topic"
}

resource "google_pubsub_subscription" "main_subscription" {
  name  = "${var.project_name}-subscription"
  topic = google_pubsub_topic.main_topic.name

  ack_deadline_seconds = 20

  retry_policy {
    minimum_backoff = "10s"
  }
}

resource "google_service_account" "cloudrun_sa" {
  account_id   = "${var.project_name}-cloudrun-sa"
  display_name = "Cloud Run Service Account"
}

resource "google_project_iam_member" "cloudrun_permissions" {
  for_each = toset([
    "roles/cloudsql.client",
    "roles/pubsub.publisher",
    "roles/pubsub.subscriber",
    "roles/storage.objectAdmin",
    "roles/secretmanager.secretAccessor",
  ])

  project = var.project_id
  role    = each.key
  member  = "serviceAccount:${google_service_account.cloudrun_sa.email}"
}

resource "google_artifact_registry_repository" "repo" {
  location      = var.region
  repository_id = "${var.project_name}-cloud-build-repo"
  format        = "DOCKER"
}

resource "google_secret_manager_secret" "secrets" {
  for_each = toset([
    "APP_ENV",
    "INFOBIP_BASE_URL",
    "INFOBIP_API_KEY",
    "JWT_SECRET_KEY",
    "JWT_ACCESS_TOKEN_EXPIRY",
    "JWT_REFRESH_TOKEN_EXPIRY",
    "DB_HOST",
    "DB_PORT",
    "DB_USER",
    "DB_PASSWORD",
    "DB_NAME",
    "DB_SSL_MODE",
    "GOOGLE_PUBSUB_PROJECT_ID",
    "GOOGLE_PUBSUB_SUBSCRIPTION_ID",
    "GOOGLE_PUBSUB_TOPIC_ID",
    "X_API_KEY",
    "GCP_BUCKET_NAME",
    "GCP_PROJECT_ID",
  ])

  secret_id = each.key

  replication {
    auto {}
  }
}

