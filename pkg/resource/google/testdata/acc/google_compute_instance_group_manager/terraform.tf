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

resource "random_string" "group-id" {
  length  = 12
  upper   = false
  special = false
}

resource "google_compute_network" "vpc_network" {
    name = "vpc-network-${random_string.net-id.result}"
}

resource "google_compute_health_check" "autohealing" {
    name                = "autohealing-health-check-${random_string.group-id.result}"
    check_interval_sec  = 5
    timeout_sec         = 5
    healthy_threshold   = 2
    unhealthy_threshold = 10 # 50 seconds

    http_health_check {
        request_path = "/healthz"
        port         = "8080"
    }
}

resource "google_compute_instance_template" "appserver" {
    name_prefix  = "instance-template-${random_string.group-id.result}-"
    machine_type = "e2-medium"
    region       = "us-central1"

    // boot disk
    disk {
        source_image      = "debian-cloud/debian-11"
        auto_delete       = true
        boot              = true
    }

    // networking
    network_interface {
        network = google_compute_network.vpc_network.name
    }

    lifecycle {
        create_before_destroy = true
    }
}

resource "google_compute_instance_group_manager" "appserver" {
    name = "appserver-igm-${random_string.group-id.result}"

    base_instance_name = "app"
    zone               = "us-central1-a"

    version {
        instance_template  = google_compute_instance_template.appserver.id
    }

    target_pools = []
    target_size  = 2

    named_port {
        name = "customhttp"
        port = 8888
    }

    auto_healing_policies {
        health_check      = google_compute_health_check.autohealing.id
        initial_delay_sec = 300
    }
}
