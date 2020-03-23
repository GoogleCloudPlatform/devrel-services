
variable "project_id" {
  description = "The GCP project ID for this module."
}

variable "mghp_bucket_name" {
  type = string
  description = "The name of the Cloud Storage Bucket to store the Private Key and Certificate in"
}

variable "mghp_certificate_secret_name" {
  type = string 
  description = "The name of the Cloud Secret Manager Secret to use for the Magic GitHub Proxy certificate"
}

variable "mghp_private_key_secret_name" {
  type = string 
  description = "The name of the Cloud Secret Manager Secret to use for the Magic GitHub Proxy private key"
}

variable "client_certificate" {}
variable "client_key" {}
variable "cluster_ca_certificate" {}
variable "host" {}