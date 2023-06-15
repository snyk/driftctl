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
  name     = "driftctl-unittest-1-${random_string.postfix.result}"
  location = "EU"
}

resource "google_storage_bucket_iam_binding" "binding_admin_1" {
  bucket = google_storage_bucket.driftctl-unittest.name
  role   = "roles/storage.admin"
  members = [
    "group:cloud-context-team@snyk.io",
  ]
}

resource "google_storage_bucket_iam_binding" "binding_viewer_1" {
  bucket = google_storage_bucket.driftctl-unittest.name
  role   = "roles/storage.objectViewer"
  members = [
    "group:cloud-context-team@snyk.io",
  ]
}

resource "google_storage_bucket" "driftctl-unittest2" {
  name     = "driftctl-unittest-2-${random_string.postfix.result}"
  location = "EU"
}

resource "google_storage_bucket_iam_binding" "binding_admin_2" {
  bucket = google_storage_bucket.driftctl-unittest2.name
  role   = "roles/storage.admin"
  members = [
    "group:cloud-context-team@snyk.io",
  ]
}
