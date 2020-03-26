

resource "kubernetes_deployment" "mghp" {
  metadata {
    name = "magic-github-proxy-deployment"
    labels = {
      app = "magic-github-proxy"
    }
  }
  spec {
    replicas = 1
    strategy {
      type = "Recreate"
    }
    template {
      metadata {
        labels = {
          app = "magic-github-proxy"
        }
      }
      spec {
        volume {
          name = "gcp-sa"
          secret {
            secret_name = kubernetes_secret.mghp_sa_secret.metadata.0.name
          }
        }
        container {
          image = "gcr.io/endpoints-release/endpoints-runtime:1"
          name  = "esp"
          args = [
            "--http_port=8080", #http
            "--service=mghp.endpoints.${var.project_id}.cloud.goog",
            "--backend=http://127.0.0.1:5000",
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
        }
        container {
          name  = "magic-github-proxy"
          image = "gcr.io/${var.project_id}/magic-github-proxy:latest"
          command = [
            "python",
            "main.py",
            "--port",
            "5000",
            "--project-id",
            "${var.project_id}",
            "--kms-location",
            "${google_kms_key_ring.mghp_key_ring.location}",
            "--kms-key-ring",
            "${google_kms_key_ring.mghp_key_ring.name}",
            "--kms-key",
            "${google_kms_crypto_key.mghp_crypto_key.name}",
            "--bucket-name",
            "${google_storage_bucket.mghp_bucket.name}",
            "--pri",
            "${google_storage_bucket_object.mghp_private_key.name}",
            "--cer",
            "${google_storage_bucket_object.mghp_cert.name}"
          ]
          port {
            container_port = 5000
          }
          volume_mount {
            name       = "gcp-sa"
            mount_path = "/var/secrets/google"
            read_only  = true
          }
          env {
            name  = "GOOGLE_APPLICATION_CREDENTIALS"
            value = "/var/secrets/google/key.json"
          }
        }
      }
    }
  }
}


resource "kubernetes_service" "mghp_np" {
  metadata {
    name = "esp-mghp-rtr-np"
    labels = {
      app = "magic-github-proxy"
    }
  }
  spec {
    type = "NodePort"
    selector = {
      app = kubernetes_deployment.mghp.metadata.0.labels.app
    }
    port {
      port        = 8080
      target_port = 8080
      name        = "http"
    }
  }
}

resource "kubernetes_ingress" "esp_samplrd_ingress" {
  metadata {
    name = "esp-mghp-ingress"
    labels = {
      app = "magic-github-proxy"
    }
    annotations = {
      "kubernetes.io/ingress.global-static-ip-name" = google_compute_global_address.mghp_ip.name
      "ingress.gcp.kubernetes.io/pre-shared-cert"   = google_compute_managed_ssl_certificate.mghp_ssl.name
    }
  }
  spec {
    backend {
      service_name = kubernetes_service.mghp_np.metadata.0.name
      service_port = 8080
    }
  }
}

### Secrets

resource "kubernetes_secret" "mghp_sa_secret" {
  metadata {
    name = "service-account-magic-github-proxy"
  }
  data = {
    "key.json" = base64decode(google_service_account_key.mghp_service_account_key.private_key)
  }
}