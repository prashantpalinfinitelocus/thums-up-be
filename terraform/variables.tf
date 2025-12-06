variable "project_id" {
  description = "GCP Project ID"
  type        = string
}

variable "project_name" {
  description = "Project name for resource naming"
  type        = string
  default     = "thums-up"
}

variable "region" {
  description = "GCP Region"
  type        = string
  default     = "asia-south1"
}

variable "database_name" {
  description = "Database name"
  type        = string
  default     = "thums-up"
}

variable "database_user" {
  description = "Database user"
  type        = string
  default     = "prashantpal"
}

variable "database_password" {
  description = "Database password"
  type        = string
  sensitive   = true
}

