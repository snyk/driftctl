provider "google" {}

terraform {
    required_version = "~> 0.15.0"
    required_providers {
        google = {
            version = "3.78.0"
        }
    }
}

resource "google_dns_managed_zone" "example-zone" {
    name        = "example-zone"
    dns_name    = "example-${random_id.rnd.hex}.com."
    description = "Example DNS zone"
    labels = {
        foo = "bar"
    }
}

resource "random_id" "rnd" {
    byte_length = 4
}
