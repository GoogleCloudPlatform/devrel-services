
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
  role   = "projects/${data.terraform_remote_state.common.outputs.project_id}/roles/${google_project_iam_custom_role.mghp_kms_access.role_id}"
  member = "serviceAccount:${google_service_account.mghp_service_account.email}"
  depends_on = [
    google_service_account.mghp_service_account,
    google_project_iam_custom_role.mghp_kms_access,
  ]
}
