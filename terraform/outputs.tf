output "database_connection_name" {
  description = "Database connection name"
  value       = google_sql_database_instance.postgres.connection_name
}

output "database_public_ip" {
  description = "Database public IP"
  value       = google_sql_database_instance.postgres.public_ip_address
}

output "bucket_name" {
  description = "Storage bucket name"
  value       = google_storage_bucket.bucket.name
}

output "pubsub_topic" {
  description = "PubSub topic name"
  value       = google_pubsub_topic.main_topic.name
}

output "pubsub_subscription" {
  description = "PubSub subscription name"
  value       = google_pubsub_subscription.main_subscription.name
}

output "service_account_email" {
  description = "Service account email"
  value       = google_service_account.cloudrun_sa.email
}

output "artifact_registry_repo" {
  description = "Artifact Registry repository"
  value       = google_artifact_registry_repository.repo.name
}

