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

resource "google_compute_ssl_certificate" "default" {
  name        = random_id.certificate.hex
  private_key = file("host.key")
  certificate = file("host.cert")

  lifecycle {
    create_before_destroy = true
  }
}

resource "random_id" "certificate" {
  byte_length = 4
  prefix      = "my-certificate-"

  keepers = {
    private_key = filebase64sha256("host.key")
    certificate = filebase64sha256("host.cert")
  }
}
