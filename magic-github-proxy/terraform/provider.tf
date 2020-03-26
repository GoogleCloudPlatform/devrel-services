provider "kubernetes" {
  host                   = module.project_resources.host
  client_certificate     = base64decode(module.project_resources.client_certificate)
  client_key             = base64decode(module.project_resources.client_key)
  cluster_ca_certificate = base64decode(module.project_resources.cluster_ca_certificate)
}

provider "google-beta" {
  project = var.project_id
  region  = var.region
  version = "~> 3.12.0"
  batching {
    enable_batching = false
  }
}
