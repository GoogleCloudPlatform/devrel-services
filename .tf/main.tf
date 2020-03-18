
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
    
  project_id = var.project_id
  region = var.region
  settings_bucket_name = var.settings_bucket_name
}

module "maintner" {
  source = "./maintner"

  project_id = var.project_id
  maintner_bucket_name = var.maintner_bucket_name
}

module "samplr" {
  source = "./samplr"

  project_id = var.project_id
}

module "mghp" {
  source = "./mghp"

  project_id = var.project_id
}