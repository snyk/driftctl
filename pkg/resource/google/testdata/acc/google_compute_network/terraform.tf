provider "google" {}

terraform {
  required_version = "~> 0.15.0"
  required_providers {
    google = {
      version = "3.78.0"
    }
  }
}

resource "random_string" "postfix" {
  length  = 6
  upper   = false
  special = false
}

resource "google_compute_network" "driftctl-unittest-1" {
  project = "driftctl-qa-1"
  name    = "driftctl-unittest-${random_string.postfix.result}"
}

resource "google_compute_network" "driftctl-unittest-2" {
  project                 = "driftctl-qa-1"
  name                    = "driftctl-unittest-${random_string.postfix.result}"
  auto_create_subnetworks = true
  mtu                     = 1460
  routing_mode            = "GLOBAL"
}

resource "google_compute_network" "driftctl-unittest-3" {
  project                         = "driftctl-qa-1"
  name                            = "driftctl-unittest-${random_string.postfix.result}"
  description                     = "driftctl test"
  auto_create_subnetworks         = false
  mtu                             = 1500
  delete_default_routes_on_create = true
  routing_mode                    = "REGIONAL"
}
