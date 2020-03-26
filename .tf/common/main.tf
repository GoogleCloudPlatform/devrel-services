#
# GKE Cluster Setup
#

resource "google_container_cluster" "devrel_services" {
  name     = "devrel-services"
  location = var.region

  # We can't create a cluster with no node pool defined, but we want to only use
  # separately managed node pools. So we create the smallest possible default
  # node pool and immediately delete it.
  remove_default_node_pool = true
  initial_node_count       = 1
}


resource "google_container_node_pool" "primary_nodes" {
  name               = "adjust-scope"
  location           = var.region
  cluster            = google_container_cluster.devrel_services.name
  initial_node_count = 10

  node_config {
    preemptible  = false
    machine_type = "n1-standard-2"

    metadata = {
      disable-legacy-endpoints = "true"
    }

    oauth_scopes = [
      "https://www.googleapis.com/auth/devstorage.read_only",
      "https://www.googleapis.com/auth/servicecontrol",
      "https://www.googleapis.com/auth/service.management.readonly",
      "https://www.googleapis.com/auth/logging.write",
      "https://www.googleapis.com/auth/monitoring",
    ]
  }

  autoscaling {
    min_node_count = 6
    max_node_count = 40
  }

  management {
    auto_repair  = true
    auto_upgrade = false
  }
}

resource "google_container_node_pool" "samplr_nodes" {
  name               = "samplr-nodes"
  location           = var.region
  cluster            = google_container_cluster.devrel_services.name
  initial_node_count = 1

  node_config {
    preemptible  = false
    machine_type = "n1-standard-2"

    metadata = {
      disable-legacy-endpoints = "true"
      # The idea of this is to pair it with a pod affinity label
      # for the samplr pods. This way we can keep samplr and maintner
      # application pods in different pools to help with node upgrades etc.
      drghs-node-type = "samplr"
    }

    oauth_scopes = [
      "https://www.googleapis.com/auth/devstorage.read_only",
      "https://www.googleapis.com/auth/servicecontrol",
      "https://www.googleapis.com/auth/service.management.readonly",
      "https://www.googleapis.com/auth/logging.write",
      "https://www.googleapis.com/auth/monitoring",
    ]
  }

  autoscaling {
    min_node_count = 1
    max_node_count = 40
  }

  management {
    auto_repair  = true
    auto_upgrade = false
  }
}

resource "google_storage_bucket" "repos_list_bucket" {
  name     = var.settings_bucket_name
  location = "US"
}
