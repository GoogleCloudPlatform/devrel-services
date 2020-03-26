
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
  protoc_output_base64 = filebase64("../../drghs/v1/api_descriptor.pb")

  depends_on = [
    data.google_compute_global_address.maintner_address,
  ]

  lifecycle {
    prevent_destroy = true
  }
}
