
resource "kubernetes_deployment" "maintner_rtr" {
  metadata {
    name = "maintnerd-rtr"
    labels = {
      app = "maintnerd-rtr"
    }
  }
  spec {
    replicas = 1
    strategy {
      type = "Recreate"
    }
    selector {
      match_labels = {
        app = "maintnerd-rtr"
      }
    }

    template {
      metadata {
        labels = {
          app = "maintnerd-rtr"
        }
      }
      spec {
        volume {
          name = "gcp-sa"
          secret {
            secret_name = kubernetes_secret.maintner_sa_secret.metadata.0.name
          }
        }
        container {
          image = "gcr.io/endpoints-release/endpoints-runtime:1"
          name  = "esp"
          args = [
            "--http_port", "8080",
            "--backend", "grpc://127.0.0.1:80",
            "--service=drghs.endpoints.${data.terraform_remote_state.common.outputs.project_id}.cloud.goog",
            "--version=${google_endpoints_service.maintner_grpc_service.config_id}",
            "--healthz=_healthz"
          ]
          port {
            container_port = "8080"
          }
          port {
            container_port = "80"
          }
          resources {
            limits {
              cpu    = "200m"
              memory = "100Mi"
            }
            requests {
              cpu    = "100m"
              memory = "50Mi"
            }
          }
          readiness_probe {
            http_get {
              path = "/_healthz"
              port = 8080
            }
            initial_delay_seconds = 30
          }
        }

        container {
          image = "gcr.io/${data.terraform_remote_state.common.outputs.project_id}/maintner-rtr:${var.image_tag}"
          name  = "maintnerd-rtr"
          command = [
            "/maintner-rtr",
            "--listen=:80",
            "--verbose",
            "--sprvsr=${kubernetes_service.maintner_sprvsr_cip.metadata.0.name}",
            "--settings-bucket=${var.settings_bucket_name}",
            "--repos-file=${var.repos_file_name}",
          ]
          port {
            container_port = 80
          }
          resources {
            limits {
              cpu    = "200m"
              memory = "100Mi"
            }
            requests {
              cpu    = "100m"
              memory = "50Mi"
            }
          }
          env {
            name  = "GOOGLE_APPLICATION_CREDENTIALS"
            value = "/var/secrets/google/key.json"
          }
          volume_mount {
            name       = "gcp-sa"
            mount_path = "/var/secrets/google"
            read_only  = true
          }
          readiness_probe {
            exec {
              command = [
                "/bin/grpc_health_probe",
                "--addr=:80"
              ]
            }
            initial_delay_seconds = 10
            period_seconds        = 3
          }
          liveness_probe {
            exec {
              command = [
                "/bin/grpc_health_probe",
                "--addr=:80"
              ]
            }
            initial_delay_seconds = 10
            period_seconds        = 3
          }
        }
      }
    }
  }
}


resource "kubernetes_secret" "maintner_sa_secret" {
  metadata {
    name = "maintner-sa"
  }
  data = {
    "key.json" = base64decode(google_service_account_key.maintner_service_account_key.private_key)
  }
}

resource "kubernetes_service" "maintner_rtr_np" {
  metadata {
    name = "esp-maintnerd-rtr-np"
  }
  spec {
    type = "NodePort"
    selector = {
      app = kubernetes_deployment.maintner_rtr.metadata.0.labels.app
    }
    port {
      port        = 80
      target_port = 8080
      name        = "http"
    }
    port {
      port        = 5000
      target_port = 80
      name        = "grpc"
    }
  }
}

resource "kubernetes_ingress" "esp_maintnerd_ingress" {
  metadata {
    name = "esp-maintnerd-ingress"
    labels = {
      app = "maintnerd"
    }
    annotations = {
      "ingress.gcp.kubernetes.io/pre-shared-cert"   = google_compute_managed_ssl_certificate.maintner-ssl.name
      "kubernetes.io/ingress.global-static-ip-name" = google_compute_global_address.maintner_ip.name
    }
  }
  spec {
    backend {
      service_name = kubernetes_service.maintner_rtr_np.metadata.0.name
      service_port = 80
    }
  }
}


resource "kubernetes_service" "maintner_sprvsr_cip" {
  metadata {
    name = "maintnerd-sprvsr-cip"
  }
  spec {
    type = "ClusterIP"
    selector = {
      app = kubernetes_deployment.maintner_sprvsr.metadata.0.labels.app
    }
    port {
      port        = 80
      target_port = 80
      name        = "http"
    }
  }
}

resource "kubernetes_deployment" "maintner_sprvsr" {
  metadata {
    name = "maintnerd-sprvsr"
    labels = {
      app = "maintnerd-sprvsr"
    }
  }
  spec {
    replicas = 1
    strategy {
      type = "Recreate"
    }
    selector {
      match_labels = {
        app = "maintnerd-sprvsr"
      }
    }
    template {
      metadata {
        labels = {
          app = "maintnerd-sprvsr"
        }
      }
      spec {
        volume {
          name = "gcp-sa"
          secret {
            secret_name = kubernetes_secret.maintner_sa_secret.metadata.0.name
          }
        }
        container {
          image = "gcr.io/${data.terraform_remote_state.common.outputs.project_id}/maintnerd-sprvsr:${var.image_tag}"
          name  = "maintnerd-sprvsr"
          command = [
            "/maintner-sprvsr",
            "--listen=:80",
            "--verbose",
            "--gcp-project=${data.terraform_remote_state.common.outputs.project_id}",
            "--github-secret=${kubernetes_secret.github_tokens.metadata.0.name}",
            "--settings-bucket=${var.settings_bucket_name}",
            "--repos-file=${var.repos_file_name}",
            "--service-account-secret=SERVICE_ACCOUNT_SECRET_NAME",
            "--maint-image-name=gcr.io/${data.terraform_remote_state.common.outputs.project_id}/maintnerd:${var.image_tag}",
            "--mutation-bucket=PREFIX",
          ]
          port {
            container_port = 80
          }

          env {
            name  = "GOOGLE_APPLICATION_CREDENTIALS"
            value = "/var/secrets/google/key.json"
          }
          volume_mount {
            name       = "gcp-sa"
            mount_path = "/var/secrets/google"
            read_only  = true
          }
          liveness_probe {
            http_get {
              path = "/_healthz"
              port = 80
            }
            initial_delay_seconds = 10
            period_seconds        = 3
          }
        }
      }
    }
  }
}

resource "kubernetes_cron_job" "sweeper" {
  metadata {
    name = "sweeper"
  }
  spec {
    # Run Every day at 02:00 Hours
    schedule = "0 2 * * *"
    job_template {
      metadata {}
      spec {
        template {
          metadata {}
          spec {
            container {
              name  = "maintner-swpr"
              image = "gcr.io/${data.terraform_remote_state.common.outputs.project_id}/maintner-swpr:${var.image_tag}"
              args = [
                "--rtr-address=${kubernetes_service.maintner_rtr_np.metadata.0.name}:5000"
              ]
              env {
                name = "GITHUB_TOKEN"
                value_from {
                  secret_key_ref {
                    key  = kubernetes_secret.sweeper_secret.metadata.0.name
                    name = keys(kubernetes_secret.sweeper_secret.data).0
                  }
                }
              }
            }
          }
        }
      }
    }
  }
}


# Pull api keys from Secret Manager to use in the 
data "google_secret_manager_secret_version" "sct" {
  for_each = var.github_api_key_secret_names

  provider = google-beta
  secret   = each.key
}

resource "kubernetes_secret" "github_tokens" {
  metadata {
    name = "github-tokens"
  }
  data = {
    for s in data.google_secret_manager_secret_version.sct :
    s.secret => s.secret_data
  }
}


# Pull api keys from Secret Manager to use in Sweeper
data "google_secret_manager_secret_version" "swpr_secret" {
  provider = google-beta
  secret   = var.sweeper_github_secret_key
}

resource "kubernetes_secret" "sweeper_secret" {
  metadata {
    name = "sweeper-secret"
  }
  data = {
    "api_key" = data.google_secret_manager_secret_version.swpr_secret.secret_data
  }
}
