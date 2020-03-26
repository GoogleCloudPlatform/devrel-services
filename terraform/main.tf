
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