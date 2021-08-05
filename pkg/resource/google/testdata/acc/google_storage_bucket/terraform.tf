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

resource "google_storage_bucket" "driftctl-unittest" {
    name          = "driftctl-unittest-${count.index}-${random_string.postfix.result}"
    location      = "EU"
    count = 3
}
