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

resource "google_storage_bucket_iam_member" "elie1" {
  bucket = google_storage_bucket.driftctl-unittest.name
  role   = "roles/storage.admin"
  member = "group:cloud-context-team@snyk.io"
}

resource "google_storage_bucket_iam_member" "will1" {
  bucket = google_storage_bucket.driftctl-unittest.name
  role   = "roles/storage.objectViewer"
  member = "group:cloud-context-team@snyk.io"
}

resource "google_storage_bucket" "driftctl-unittest2" {
  name     = "driftctl-unittest-2-${random_string.postfix.result}"
  location = "EU"
}

resource "google_storage_bucket_iam_member" "eli2" {
  bucket = google_storage_bucket.driftctl-unittest2.name
  role   = "roles/storage.objectViewer"
  member = "group:cloud-context-team@snyk.io"
}

resource "google_storage_bucket_iam_member" "will2" {
  bucket = google_storage_bucket.driftctl-unittest2.name
  role   = "roles/storage.admin"
  member = "group:cloud-context-team@snyk.io"
}
