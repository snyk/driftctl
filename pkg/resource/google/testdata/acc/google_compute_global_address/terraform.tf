provider "google" {}

terraform {
    required_version = "~> 0.15.0"
    required_providers {
        google = {
            version = "3.78.0"
        }
    }
}

resource "google_compute_global_address" "default" {
    name = "global-appserver-ip"
}

resource "google_compute_address" "ip_address" {
    name = "my-address"
    region = "us-central1"
}
