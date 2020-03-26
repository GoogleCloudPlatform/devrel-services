#
# Samplr Service Account
#

resource "google_service_account" "samplr_service_account" {
  account_id   = "samplr"
  display_name = "Samplr Service Account"
  description  = "Service Account used by Samplr service"
}


resource "google_service_account_key" "samplr_service_account_key" {
  service_account_id = google_service_account.samplr_service_account.name
}

data "google_service_account_key" "samplr_service_account_key" {
  name = google_service_account_key.samplr_service_account_key.name
}

resource "google_storage_bucket_iam_binding" "editor" {
  bucket = var.settings_bucket_name
  role   = "roles/storage.admin"
  members = [
    "serviceAccount:${google_service_account.samplr_service_account.email}",
  ]
}

resource "google_project_iam_member" "error_reporting" {
  project = var.project_id
  role    = "roles/errorreporting.writer"
  member  = "serviceAccount:${google_service_account.samplr_service_account.email}"
}