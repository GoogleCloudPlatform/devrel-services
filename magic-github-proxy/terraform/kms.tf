
resource "google_kms_key_ring" "mghp_key_ring" {
  name     = "magic-github-proxy"
  location = "global"
}

resource "google_kms_crypto_key" "mghp_crypto_key" {
  name     = "enc-at-rest"
  key_ring = google_kms_key_ring.mghp_key_ring.self_link
  purpose  = "ENCRYPT_DECRYPT"
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
