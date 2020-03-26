resource "kubernetes_deployment" "samplr_rtr" {
  metadata {
    name = "samplrd-rtr"
    labels = {
      app = "samplrd-rtr"
    }
  }
  spec {
    replicas = 1
    strategy {
      type = "Recreate"
    }
    selector {
      match_labels = {
        app = "samplrd-rtr"
      }
    }
    template {
      metadata {
        labels = {
          app = "samplrd-rtr"
        }
      }
      spec {
        volume {
          name = "gcp-sa"
          secret {
            secret_name = kubernetes_secret.samplr_sa_secret.metadata.0.name
          }
        }
        container {
          image = "gcr.io/endpoints-release/endpoints-runtime:1"
          name  = "esp"
          args = [
            "--http_port", "8080",
            "--backend", "grpc://127.0.0.1:80",
            "--service=samplr.endpoints.${var.project_id}.cloud.goog",
            "--version=${google_endpoints_service.samplr_grpc_service.config_id}",
            "--healthz=_healthz"
          ]
          port {
            container_port = "8080"
          }
          readiness_probe {
            http_get {
              path = "/_healthz"
              port = 8080
            }
            initial_delay_seconds = 30
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
        }
        container {
          image = "gcr.io/${var.project_id}/samplr-rtr:latest"
          name  = "samplrd-rtr"
          command = [
            "/samplr-rtr",
            "--listen=:80",
            "--verbose",
          ]
          port {
            container_port = "80"
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
            period_seconds        = 10
          }
          liveness_probe {
            exec {
              command = [
                "/bin/grpc_health_probe",
                "--addr=:80"
              ]
            }
            initial_delay_seconds = 5
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
        }
      }
    }
  }
}

resource "kubernetes_service" "samplr_rtr_np" {
  metadata {
    name = "esp-samplrd-rtr-np"
  }
  spec {
    type = "NodePort"
    selector = {
      app = kubernetes_deployment.samplr_rtr.metadata.0.labels.app
    }
    port {
      port        = 80
      target_port = 8080
      name        = "http"
    }
  }
}


resource "kubernetes_ingress" "esp_samplrd_ingress" {
  metadata {
    name = "esp-samplrd-ingress"
    labels = {
      app = "samplrd"
    }
    annotations = {
      "kubernetes.io/ingress.global-static-ip-name" = google_compute_global_address.samplr_ip.name
      "ingress.gcp.kubernetes.io/pre-shared-cert"   = google_compute_managed_ssl_certificate.samplr-ssl.name
    }
  }
  spec {
    backend {
      service_name = kubernetes_service.samplr_rtr_np.metadata.0.name
      service_port = 80
    }
  }
}


resource "kubernetes_deployment" "samplr_sprvsr" {
  metadata {
    name = "samplrd-sprvsr"
    labels = {
      app = "samplrd-sprvsr"
    }
  }
  spec {
    replicas = 1
    strategy {
      type = "Recreate"
    }
    selector {
      match_labels = {
        app = "samplrd-sprvsr"
      }
    }
    template {
      metadata {
        labels = {
          app = "samplrd-sprvsr"
        }
      }
      spec {
        service_account_name            = kubernetes_service_account.samplr_sprvsr_sa.metadata.0.name
        automount_service_account_token = true
        volume {
          name = "gcp-sa"
          secret {
            secret_name = kubernetes_secret.samplr_sa_secret.metadata.0.name
          }
        }
        container {
          name  = "samplrd-sprvsr"
          image = "gcr.io/${var.project_id}/samplr-sprvsr:latest"
          command = [
            "/samplr-sprvsr",
            "--listen=:80",
            "--verbose",
            "--gcp-project=${var.project_id}",
            "--settings-bucket=${var.settings_bucket_name}",
            "--repos-file=${var.repos_file_name}",
            "--service-account-secret=SERVICE_ACCOUNT_SECRET_NAME",
            "--samplr-image-name=gcr.io/${var.project_id}/samplrd:latest"
          ]
          liveness_probe {
            http_get {
              path = "/_healthz"
              port = 80
            }
            initial_delay_seconds = 10
            period_seconds        = 3
          }
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
        }
      }
    }
  }
}

resource "kubernetes_service" "samplrd_sprvsr_cip" {
  metadata {
    name = "samplrd-sprvsr-cip"
  }
  spec {
    type = "ClusterIP"
    selector = {
      app = kubernetes_deployment.samplr_sprvsr.metadata.0.labels.app
    }
    port {
      port        = 80
      target_port = 80
      name        = "http"
    }
  }
}

resource "kubernetes_secret" "samplr_sa_secret" {
  metadata {
    name = "samplr-sa"
  }
  data = {
    "key.json" = base64decode(google_service_account_key.samplr_service_account_key.private_key)
  }
}

resource "kubernetes_service_account" "samplr_sprvsr_sa" {
  metadata {
    name = "samplr-sprvsr-sa"
  }
  secret {
    name = "samplr-sprvsr-sa"
  }
}

resource "kubernetes_cluster_role_binding" "samplr_sprvsr_sa_edit" {
  metadata {
    name = "samplr-sprvsr-edit"
  }
  role_ref {
    kind      = "ClusterRole"
    name      = "edit"
    api_group = "rbac.authorization.k8s.io"
  }
  subject {
    kind      = "ServiceAccount"
    name      = kubernetes_service_account.samplr_sprvsr_sa.metadata.0.name
    namespace = "default"
  }
}