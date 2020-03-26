
resource "google_compute_global_address" "maintner_ip" {
  name = "maintner-ip"
}

data "google_compute_global_address" "maintner_address" {
  name = "maintner-ip"
  depends_on = [
    google_compute_global_address.maintner_ip,
  ]
}


resource "google_compute_managed_ssl_certificate" "maintner-ssl" {
  provider = google-beta

  name = "drghs-endpoints-cert"

  managed {
    domains = [google_endpoints_service.maintner_grpc_service.service_name]
  }
}
