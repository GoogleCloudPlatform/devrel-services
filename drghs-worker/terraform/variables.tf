variable "core_state_bucket" {
  type        = string
  description = "The name of the GCS bucket which stores the state of the core infrastructure"
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

variable "image_tag" {
  type        = string
  description = "The tag of the Docker Images to use to deploy"
  default     = "latest"
}
