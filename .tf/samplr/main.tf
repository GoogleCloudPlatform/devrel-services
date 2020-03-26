#
# Samplr Cloud Endpoints
#

resource "google_compute_global_address" "samplr_ip" {
  name = "samplr-ip"
}

data "google_compute_global_address" "samplr_address" {
  name = "samplr-ip"

  depends_on = [
    google_compute_global_address.samplr_ip
  ]
}

resource "google_endpoints_service" "samplr_grpc_service" {
  service_name         = "samplr.endpoints.${var.project_id}.cloud.goog"
  grpc_config          = <<-EOT
  type: google.api.Service
  config_version: 3

  name: samplr.endpoints.${var.project_id}.cloud.goog
  title: samplr gRPC API (TYPE)

  apis:
  - name: drghs.v1.SampleService

  endpoints:
  - name: samplr.endpoints.${var.project_id}.cloud.goog
    target: "${data.google_compute_global_address.samplr_address.address}"
  EOT
  protoc_output_base64 = filebase64("../drghs/v1/api_descriptor.pb")

  depends_on = [
    data.google_compute_global_address.samplr_address,
  ]

  lifecycle {
    prevent_destroy = true
  }
}

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

#
# SSL Cert
#

resource "google_compute_managed_ssl_certificate" "samplr-ssl" {
  provider = google-beta
  name     = "samplr-endpoints-cert"
  managed {
    domains = [google_endpoints_service.samplr_grpc_service.service_name]
  }
}