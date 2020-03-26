# Provider definitions

provider "google" {
  project = var.project_id
  region  = var.region
  version = "~> 3.12.0"
  batching {
    enable_batching = false
  }
}

provider "google-beta" {
  project = var.project_id
  region  = var.region
  version = "~> 3.12.0"
  batching {
    enable_batching = false
  }
}
