# Variables for the Platform/Core

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

#  Variables for MGHP
variable "mghp_bucket_name" {
  type        = string
  description = "The name of the Cloud Storage Bucket to store the Private Key and Certificate in"
}

variable "mghp_certificate_secret_name" {
  type        = string
  description = "The name of the Cloud Secret Manager Secret to use for the Magic GitHub Proxy certificate"
}

variable "mghp_private_key_secret_name" {
  type        = string
  description = "The name of the Cloud Secret Manager Secret to use for the Magic GitHub Proxy private key"
}
