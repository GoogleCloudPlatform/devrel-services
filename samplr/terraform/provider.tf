provider "kubernetes" {
  host                   = data.terraform_remote_state.common.outputs.host
  client_certificate     = base64decode(data.terraform_remote_state.common.outputs.client_certificate)
  client_key             = base64decode(data.terraform_remote_state.common.outputs.client_key)
  cluster_ca_certificate = base64decode(data.terraform_remote_state.common.outputs.cluster_ca_certificate)
}

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
