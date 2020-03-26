
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
  protoc_output_base64 = filebase64("../../drghs/v1/api_descriptor.pb")

  depends_on = [
    data.google_compute_global_address.samplr_address,
  ]

  lifecycle {
    prevent_destroy = true
  }
}

#