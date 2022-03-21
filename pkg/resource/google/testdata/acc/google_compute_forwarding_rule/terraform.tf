provider "google" {
  region  = "us-east1"
}

terraform {
  required_version = "~> 0.15.0"
  required_providers {
    google = {
      version = "3.78.0"
    }
  }
}

resource "google_compute_forwarding_rule" "default" {
    name       = "foo-forwarding-rule"
    target     = google_compute_target_pool.default.id
    port_range = "80"
}

resource "google_compute_target_pool" "default" {
    name = "foo-target-pool"
}
