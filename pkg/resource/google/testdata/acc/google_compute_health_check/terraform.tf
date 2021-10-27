provider "google" {}

terraform {
    required_version = "~> 0.15.0"
    required_providers {
        google = {
            version = "3.78.0"
        }
    }
}

resource "google_compute_health_check" "https-health-check" {
    name        = "https-health-check"
    description = "Health check via https"

    timeout_sec         = 1
    check_interval_sec  = 3
    healthy_threshold   = 4
    unhealthy_threshold = 5

    https_health_check {
        port_name          = "health-check-port"
        port_specification = "USE_NAMED_PORT"
        host               = "google.com"
        request_path       = "/"
        proxy_header       = "NONE"
        response           = "I AM HEALTHY"
    }
}
