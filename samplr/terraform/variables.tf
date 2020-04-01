variable "repos_file_name" {
  description = "The name of the file which lists the repositories to track"
  default     = "public_repos.json"
}

variable "core_state_bucket" {
  type        = string
  description = "The name of the GCS bucket which stores the state of the core infrastructure"
}

variable "image_tag" {
  type        = string
  description = "The tag of the Docker Images to use to deploy"
  default     = "latest"
}
