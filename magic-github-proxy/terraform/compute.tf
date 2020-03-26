
#
# Magic GitHub Proxy Endpoints Configuration
#

resource "google_compute_global_address" "mghp_ip" {
  name = "magic-github-proxy-ip"
}

data "google_compute_global_address" "mghp_address" {
  name = "magic-github-proxy-ip"
  depends_on = [
    google_compute_global_address.mghp_ip,
  ]
}


#
# SSL Cert
#


resource "google_compute_managed_ssl_certificate" "mghp_ssl" {
  provider = google-beta
  name     = "mghp-endpoints-cert"
  managed {
    domains = [google_endpoints_service.mghp_service.service_name]
  }
}
