provider "google" {}

terraform {
  required_version = "~> 0.15.0"
  required_providers {
    google = {
      version = "3.78.0"
    }
  }
}

resource "google_compute_network" "default" {
  name = "compute-firewall"
}

resource "google_compute_router" "router-0" {
  name    = "test-router-0"
  region  = "us-central1"
  network = google_compute_network.default.name
  bgp {
    asn               = 64514
    advertise_mode    = "CUSTOM"
    advertised_groups = ["ALL_SUBNETS"]
    advertised_ip_ranges {
      range = "1.2.3.4"
    }
    advertised_ip_ranges {
      range = "6.7.0.0/16"
    }
  }
}
