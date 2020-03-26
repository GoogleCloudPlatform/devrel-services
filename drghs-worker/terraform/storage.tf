
resource "google_storage_bucket" "maintner_bucket" {
  name     = var.maintner_bucket_name
  location = "US"
}
