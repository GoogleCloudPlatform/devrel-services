
variable "project_id" {
  description = "The GCP project ID for this module."
}


variable "repos_file_name" {
  description = "The name of the file which lists the repositories to track"
  default     = "public_repos.json"
}

variable "settings_bucket_name" {
  description = "The name of the bucket that stores the list of repositories to track"
}

variable "client_certificate" {}
variable "client_key" {}
variable "cluster_ca_certificate" {}
variable "host" {}