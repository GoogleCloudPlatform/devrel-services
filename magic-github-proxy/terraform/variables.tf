variable "core_state_bucket" {
  type        = string
  description = "The name of the GCS bucket which stores the state of the core infrastructure"
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

variable "image_tag" {
  type        = string
  description = "The tag of the Docker Images to use to deploy"
  default     = "latest"
}
