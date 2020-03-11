
#
# Maintner Cloud Endpoints
#

resource "google_compute_global_address" "maintner_ip" {
  name = "maintner-ip"
}

data "google_compute_global_address" "maintner_address" {
  name = "maintner-ip"
  depends_on = [
    google_compute_global_address.maintner_ip,
  ]
}

resource "google_endpoints_service" "maintner_grpc_service" {
  service_name         = "drghs.endpoints.${var.project_id}.cloud.goog"
  grpc_config          = <<-EOT
  type: google.api.Service
  config_version: 3

  name: drghs.endpoints.${var.project_id}.cloud.goog
  title: DevRel GitHub Services API (TYPE)

  apis:
  - name: drghs.v1.IssueService
  - name: drghs.v1.IssueServiceAdmin

  endpoints:
  - name: drghs.endpoints.${var.project_id}.cloud.goog
    target: "${data.google_compute_global_address.maintner_address.address}"
  EOT
  protoc_output_base64 = filebase64("../drghs/v1/api_descriptor.pb")

  depends_on = [
    data.google_compute_global_address.maintner_address,
  ]

  lifecycle {
    prevent_destroy = true
  }
}


resource "google_storage_bucket" "maintner_bucket" {
  name     = var.maintner_bucket_name
  location = "US"
}

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
  role   = "projects/${var.project_id}/roles/${google_project_iam_custom_role.maintner_sprvsr_bucket_creator.role_id}"
  member = "serviceAccount:${google_service_account.maintner_service_account.email}"
}
