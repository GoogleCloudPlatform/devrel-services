
resource "google_compute_global_address" "samplr_ip" {
  name = "samplr-ip"
}

data "google_compute_global_address" "samplr_address" {
  name = "samplr-ip"

  depends_on = [
    google_compute_global_address.samplr_ip
  ]
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