output "cluster_name" {
    value = google_container_cluster.devrel_services.name
}

output "client_certificate" {
  value = "${google_container_cluster.devrel_services.master_auth.0.client_certificate}"
  sensitive = true
}

output "client_key" {
  value = "${google_container_cluster.devrel_services.master_auth.0.client_key}"
  sensitive = true
}


output "cluster_ca_certificate" {
  value = "${google_container_cluster.devrel_services.master_auth.0.cluster_ca_certificate}"
  sensitive = true
}


output "host" {
  value = "${google_container_cluster.devrel_services.endpoint}"
  sensitive = true
}

output "settings_bucket_name" {
  value = "${google_storage_bucket.repos_list_bucket.name}"
}

