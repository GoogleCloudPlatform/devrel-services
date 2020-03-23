
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

resource "google_endpoints_service" "mghp_service" {
  service_name         = "magic-github-proxy.endpoints.${var.project_id}.cloud.goog"
  openapi_config       = templatefile("${path.module}/magic-github-proxy.yaml.tmpl",
  {
    project_id = var.project_id,
    ip_addr = data.google_compute_global_address.mghp_address.address,  
  })
  
  depends_on = [
    data.google_compute_global_address.mghp_address,
  ]

  lifecycle {
    prevent_destroy = true
  }
}

#
# Magic GitHubProxy Service Account
#
resource "google_service_account" "mghp_service_account" {
  account_id   = "magic-github-proxy"
  display_name = "Magic Github Proxy Account"
  description  = "Service Account used by Magic GitHub Proxy service"
}

resource "google_service_account_key" "mghp_service_account_key" {
  service_account_id = google_service_account.mghp_service_account.name
}

data "google_service_account_key" "mghp_service_account_key" {
  name = google_service_account_key.mghp_service_account_key.name
}


resource "google_project_iam_custom_role" "mghp_kms_access" {
  role_id     = "magic_github_proxy_kms_accessor"
  title       = "Magic GitHub Proxy KMS"
  description = "Allows Access to Magic Github Proxy Keys"
  permissions = [
    "cloudkms.cryptoKeyVersions.useToDecrypt"
  ]
}

resource "google_project_iam_member" "mghp_kms_iam" {
  role   = "projects/${var.project_id}/roles/${google_project_iam_custom_role.mghp_kms_access.role_id}"
  member = "serviceAccount:${google_service_account.mghp_service_account.email}"
  depends_on = [
    google_service_account.mghp_service_account,
    google_project_iam_custom_role.mghp_kms_access,
  ]
}

resource "google_kms_key_ring" "mghp_key_ring"{
  name = "magic-github-proxy"
  location = "global"
}

resource "google_kms_crypto_key" "mghp_crypto_key" {
  name     = "enc-at-rest"
  key_ring = google_kms_key_ring.mghp_key_ring.self_link
  purpose = "ENCRYPT_DECRYPT"
  depends_on = [
    google_kms_key_ring.mghp_key_ring,
  ]
  lifecycle {
    prevent_destroy = true
  }
}

resource "google_kms_crypto_key_iam_member" "mghp_crypto_key_decrypter" {
  crypto_key_id = google_kms_crypto_key.mghp_crypto_key.id
  role          = "roles/cloudkms.cryptoKeyDecrypter"
  member        = "serviceAccount:${google_service_account.mghp_service_account.email}"
  depends_on = [
        google_service_account.mghp_service_account,
        google_kms_crypto_key.mghp_crypto_key
  ]
}

#
# SSL Cert
#


resource "google_compute_managed_ssl_certificate" "mghp_ssl" {
  provider = google-beta
  name = "mghp-endpoints-cert"
  managed {
    domains = [google_endpoints_service.mghp_service.service_name]
  }
}

#
# Private Key and Certificate
#

#
# MGHP BUCKET
#
resource "google_storage_bucket" "mghp_bucket" {
  name = var.mghp_bucket_name
  location = "US"
}

data "google_secret_manager_secret_version" "mghp_private_key" {
    provider = google-beta
    secret = var.mghp_private_key_secret_name
}

resource "google_storage_bucket_object" "mghp_private_key" {
  name   = "private.pem.enc"
  content = data.google_secret_manager_secret_version.mghp_private_key.secret_data
  bucket = google_storage_bucket.mghp_bucket.name
}

data "google_secret_manager_secret_version" "mghp_certificate" {
    provider = google-beta
    secret = var.mghp_certificate_secret_name
}

resource "google_storage_bucket_object" "mghp_cert" {
  name = "public.x509.cer.enc"
  content = data.google_secret_manager_secret_version.mghp_certificate.secret_data
  bucket = google_storage_bucket.mghp_bucket.name
}
