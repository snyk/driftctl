provider "google" {}

terraform {
  required_version = "~> 0.15.0"
  required_providers {
    google = {
      version = "3.78.0"
    }
  }
}

resource "google_cloud_run_service" "default" {
    name     = "cloudrun-srv"
    location = "us-central1"

    template {
        spec {
            containers {
                image = "us-docker.pkg.dev/cloudrun/container/hello"
            }
        }
    }

    traffic {
        percent         = 100
    }
}
