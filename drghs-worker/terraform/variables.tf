
variable "project_id" {
  description = "The GCP project ID for this module."
}

variable "project_name" {
  description = "The GCP project name for this module, *might* be the same as project_id."
}

variable "region" {
  description = "The GCP region for this module."
}

variable "folder_id" {
  description = "The Folder the GCP Project is stored in."
}

variable "billing_account" {
  type        = string
  description = "The Billing Account associated with the project"
}

variable "maintner_bucket_name" {
  description = "Name of the GCS bucket to store Maintner logs to"
}

variable "repos_file_name" {
  description = "The name of the file which lists the repositories to track"
  default     = "public_repos.json"
}

variable "settings_bucket_name" {
  description = "The name of the bucket that stores the list of repositories to track"
}

variable "github_api_key_secret_names" {
  type        = set(string)
  description = "List of names of Cloud Secret Manager Secrets for API keys."
}

variable "sweeper_github_secret_key" {
  type        = string
  description = "The name of the Cloud Secret Manager Secret to use for sweeper"
}
