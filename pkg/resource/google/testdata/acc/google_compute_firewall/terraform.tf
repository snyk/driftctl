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

resource "google_compute_firewall" "default" {
    count = 3
    name    = "test-firewall-${count.index}"
    network = google_compute_network.default.name

    allow {
        protocol = "icmp"
    }

    allow {
        protocol = "tcp"
        ports    = ["80", "8080", "1000-2000"]
    }

    source_tags = ["web"]
}
