provider "google" {}

resource "google_compute_network" "default" {
    name = "test-network"
}

resource "google_compute_instance_group" "test" {
    name        = "terraform-test"
    description = "Terraform test instance group"
    zone        = "us-central1-a"
    network     = google_compute_network.default.id
}
