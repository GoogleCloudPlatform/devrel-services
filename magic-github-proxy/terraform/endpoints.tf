
resource "google_endpoints_service" "mghp_service" {
  service_name = "magic-github-proxy.endpoints.${var.project_id}.cloud.goog"
  openapi_config = templatefile("${path.module}/magic-github-proxy.yaml.tmpl",
    {
      project_id = var.project_id,
      ip_addr    = data.google_compute_global_address.mghp_address.address,
  })

  depends_on = [
    data.google_compute_global_address.mghp_address,
  ]

  lifecycle {
    prevent_destroy = true
  }
}
