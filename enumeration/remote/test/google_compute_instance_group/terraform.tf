provider "google" {}

resource "google_compute_network" "default" {
  name = "test-network"
}

resource "google_compute_instance_group" "test-1" {
  name        = "driftctl-test-1"
  description = "Terraform test instance group"
  zone        = "us-central1-a"
  network     = google_compute_network.default.id
}

resource "google_compute_instance_group" "test-2" {
  name        = "driftctl-test-2"
  description = "Terraform test instance group"
  zone        = "us-central1-a"
  network     = google_compute_network.default.id
}
