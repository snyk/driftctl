provider "google" {}

terraform {
  required_version = "~> 0.15.0"
  required_providers {
    google = {
      version = "3.78.0"
    }
  }
}

resource "random_string" "net-id" {
  length  = 12
  upper   = false
  special = false
}

resource "random_string" "subnet-id" {
  length  = 12
  upper   = false
  special = false
}

resource "google_compute_network" "default" {
  name                    = "test-network-${random_string.net-id.result}"
  auto_create_subnetworks = false
}

resource "google_compute_subnetwork" "network-with-private-secondary-ip-ranges" {
  name          = "driftctl-acc-subnet-${random_string.subnet-id.result}"
  ip_cidr_range = "10.2.0.0/16"
  network       = google_compute_network.default.id
  region        = "us-central1"
  secondary_ip_range {
    range_name    = "tf-test-secondary-range-update1"
    ip_cidr_range = "192.168.10.0/24"
  }
}
