# Provider definitions

provider "google" {
  project = var.project_id
  region  = var.region
  version = "~> 2.14"
  batching {
    enable_batching = false
  }
}

provider "google-beta" {
  project = var.project_id
  region  = var.region
  version = "~> 2.20"
  batching {
    enable_batching = false
  }
}
