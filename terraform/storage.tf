resource "google_storage_bucket" "repos_list_bucket" {
  name     = var.settings_bucket_name
  location = "US"
}
