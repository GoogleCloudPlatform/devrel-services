
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

variable "repos_file_name" {
  description = "The name of the file which lists the repositories to track"
  default     = "public_repos.json"
}

variable "core_state_bucket" {
  type        = string
  description = "The name of the GCS bucket which stores the state of the core infrastructure"
}
