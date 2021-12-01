provider "google" {}

terraform {
  required_version = "~> 0.15.0"
  required_providers {
    google = {
      version = "3.78.0"
    }
  }
}

resource "google_compute_node_template" "soletenant-tmpl" {
    name      = "soletenant-tmpl"
    region    = "us-central1"
    node_type = "n1-node-96-624"
}

resource "google_compute_node_group" "simple_nodes" {
    name        = "simple-group"
    zone        = "us-central1-f"
    description = "example google_compute_node_group for Terraform Google Provider"

    size          = 1
    node_template = google_compute_node_template.soletenant-tmpl.id
}

resource "google_compute_node_group" "nodes" {
    name        = "soletenant-group"
    zone        = "us-central1-f"
    description = "example google_compute_node_group for Terraform Google Provider"
    maintenance_policy = "RESTART_IN_PLACE"
    maintenance_window {
        start_time = "08:00"
    }
    initial_size  = 1
    node_template = google_compute_node_template.soletenant-tmpl.id
    autoscaling_policy {
        mode      = "ONLY_SCALE_OUT"
        min_nodes = 1
        max_nodes = 10
    }
}
