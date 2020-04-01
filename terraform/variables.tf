# Variables for DRGHS

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

variable "settings_bucket_name" {
  description = "Name of the GCS bucket to store the list of Repositories"
}
