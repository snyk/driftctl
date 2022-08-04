provider "google" {}

terraform {
    required_version = "~> 0.15.0"
    required_providers {
        google = {
            version = "3.78.0"
        }
    }
}

resource "google_compute_network" "driftctl-unittest-instance" {
    name    = "driftctl-unittest-instance"
}

resource "google_compute_instance" "default" {
    name         = "test"
    machine_type = "e2-medium"
    zone         = "us-central1-a"

    boot_disk {
        initialize_params {
            image = "debian-cloud/debian-11"
        }
    }

    network_interface {
        network = google_compute_network.driftctl-unittest-instance.name
    }
}
