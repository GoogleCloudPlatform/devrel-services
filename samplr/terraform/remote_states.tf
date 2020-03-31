data "terraform_remote_state" "common" {
  backend = "gcs"
  config = {
    prefix = "terraform/state"
    bucket = var.core_state_bucket
  }
}
