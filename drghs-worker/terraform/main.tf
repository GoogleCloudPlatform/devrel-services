
module "project_resources" {
  source = "../../terraform"

  project_id           = var.project_id
  project_name         = var.project_name
  region               = var.region
  settings_bucket_name = var.settings_bucket_name
  billing_account      = var.billing_account
  folder_id            = var.folder_id
}
