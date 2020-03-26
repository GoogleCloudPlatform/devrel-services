#
# Private Key and Certificate
#

#
# MGHP BUCKET
#
resource "google_storage_bucket" "mghp_bucket" {
  name     = var.mghp_bucket_name
  location = "US"
}

data "google_secret_manager_secret_version" "mghp_private_key" {
  provider = google-beta
  secret   = var.mghp_private_key_secret_name
}

resource "google_storage_bucket_object" "mghp_private_key" {
  name    = "private.pem.enc"
  content = data.google_secret_manager_secret_version.mghp_private_key.secret_data
  bucket  = google_storage_bucket.mghp_bucket.name
}

data "google_secret_manager_secret_version" "mghp_certificate" {
  provider = google-beta
  secret   = var.mghp_certificate_secret_name
}

resource "google_storage_bucket_object" "mghp_cert" {
  name    = "public.x509.cer.enc"
  content = data.google_secret_manager_secret_version.mghp_certificate.secret_data
  bucket  = google_storage_bucket.mghp_bucket.name
}
