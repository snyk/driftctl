provider "google" {}

terraform {
  required_version = "~> 0.15.0"
  required_providers {
    google = {
      version = "3.78.0"
    }
  }
}

resource "random_string" "driftctl-unittest-1" {
  length  = 12
  upper   = false
  special = false
}

resource "random_string" "driftctl-unittest-2" {
  length  = 12
  upper   = false
  special = false
}

resource "random_string" "driftctl-unittest-3" {
  length  = 12
  upper   = false
  special = false
}

resource "google_compute_network" "driftctl-unittest-1" {
  name    = "driftctl-unittest-${random_string.driftctl-unittest-1.result}"
}

resource "google_compute_network" "driftctl-unittest-2" {
  name                    = "driftctl-unittest-${random_string.driftctl-unittest-2.result}"
  auto_create_subnetworks = true
  mtu                     = 1460
  routing_mode            = "GLOBAL"
}

resource "google_compute_network" "driftctl-unittest-3" {
  name                            = "driftctl-unittest-${random_string.driftctl-unittest-3.result}"
  description                     = "driftctl test"
  auto_create_subnetworks         = false
  mtu                             = 1500
  delete_default_routes_on_create = true
  routing_mode                    = "REGIONAL"
}
