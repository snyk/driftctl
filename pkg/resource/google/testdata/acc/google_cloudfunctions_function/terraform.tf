provider "google" {
    region = "us-central1"
}

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

resource "google_storage_bucket" "bucket" {
    name = "driftctl-qa-${random_string.postfix.result}"
}

resource "google_storage_bucket_object" "archive" {
    name   = "index.zip"
    bucket = google_storage_bucket.bucket.name
    source = "./index.zip"
}

resource "google_cloudfunctions_function" "function" {
    name        = "function-test"
    description = "My function"
    runtime     = "nodejs14"

    available_memory_mb   = 128
    source_archive_bucket = google_storage_bucket.bucket.name
    source_archive_object = google_storage_bucket_object.archive.name
    trigger_http          = true
    entry_point           = "helloHttp"
}
