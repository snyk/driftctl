provider "google" {
  region  = "us-east1"
}

terraform {
  required_version = "~> 0.15.0"
  required_providers {
    google = {
      version = "4.48.0"
    }
  }
}

resource "google_kms_key_ring" "keyring" {
  name     = "keyring-example"
  location = "global"
}

resource "google_kms_crypto_key" "example-key" {
  name            = "crypto-key-example"
  key_ring        = google_kms_key_ring.keyring.id
  rotation_period = "100000s"
}
