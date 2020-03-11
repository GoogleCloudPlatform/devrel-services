
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