provider "kubernetes" {
  host                   = module.project_resources.host
  client_certificate     = base64decode(module.project_resources.client_certificate)
  client_key             = base64decode(module.project_resources.client_key)
  cluster_ca_certificate = base64decode(module.project_resources.cluster_ca_certificate)
}