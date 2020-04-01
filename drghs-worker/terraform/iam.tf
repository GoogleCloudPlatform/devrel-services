
#
# Maintner Service Account
#

resource "google_project_iam_custom_role" "maintner_sprvsr_bucket_creator" {
  role_id     = "maintner_sprvsr_bucket_creator"
  title       = "Maintner Supervisor Bucket Creator"
  description = "Used by maintner-sprvsr to create buckets"
  permissions = [
    "storage.buckets.create",
    "storage.buckets.delete",
    "storage.buckets.get",
    "storage.buckets.list",
    "storage.objects.create",
    "storage.objects.delete",
    "storage.objects.list",
  ]
}


resource "google_service_account" "maintner_service_account" {
  account_id   = "maintnerd"
  display_name = "Maintnerd Service Account"
  description  = "Service Account used by Maintner service"
}

resource "google_project_iam_member" "maintner_account_iam" {
  role   = "projects/${data.terraform_remote_state.common.outputs.project_id}/roles/${google_project_iam_custom_role.maintner_sprvsr_bucket_creator.role_id}"
  member = "serviceAccount:${google_service_account.maintner_service_account.email}"
}

resource "google_project_iam_member" "error_reporting" {
  project = data.terraform_remote_state.common.outputs.project_id
  role    = "roles/errorreporting.writer"
  member  = "serviceAccount:${google_service_account.maintner_service_account.email}"
}


resource "google_service_account_key" "maintner_service_account_key" {
  service_account_id = google_service_account.maintner_service_account.name
}

data "google_service_account_key" "maintner_service_account_key" {
  name = google_service_account_key.maintner_service_account_key.name
}
