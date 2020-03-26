
resource "google_project" "devrel-services" {
  billing_account = var.billing_account
  folder_id       = var.folder_id
  name            = var.project_name
  project_id      = var.project_id
  labels = {
    env  = "prod"
    team = "cloud_devrel_infra"
  }

  lifecycle {
    prevent_destroy = true
  }
}

module "project_resources" {
  source = "./common"

  project_id           = var.project_id
  region               = var.region
  settings_bucket_name = var.settings_bucket_name
}

module "maintner" {
  source = "./maintner"

  project_id           = var.project_id
  maintner_bucket_name = var.maintner_bucket_name

  host                   = "${module.project_resources.host}"
  client_key             = "${module.project_resources.client_key}"
  client_certificate     = "${module.project_resources.client_certificate}"
  cluster_ca_certificate = "${module.project_resources.cluster_ca_certificate}"

  settings_bucket_name = "${module.project_resources.settings_bucket_name}"

  github_api_key_secret_names = var.github_api_key_secret_names
  sweeper_github_secret_key   = var.sweeper_github_secret_key
}

module "samplr" {
  source = "./samplr"

  project_id = var.project_id

  host                   = "${module.project_resources.host}"
  client_key             = "${module.project_resources.client_key}"
  client_certificate     = "${module.project_resources.client_certificate}"
  cluster_ca_certificate = "${module.project_resources.cluster_ca_certificate}"

  settings_bucket_name = "${module.project_resources.settings_bucket_name}"
}

module "mghp" {
  source = "./mghp"

  project_id = var.project_id

  host                   = "${module.project_resources.host}"
  client_key             = "${module.project_resources.client_key}"
  client_certificate     = "${module.project_resources.client_certificate}"
  cluster_ca_certificate = "${module.project_resources.cluster_ca_certificate}"

  mghp_bucket_name             = var.mghp_bucket_name
  mghp_certificate_secret_name = var.mghp_certificate_secret_name
  mghp_private_key_secret_name = var.mghp_private_key_secret_name
}