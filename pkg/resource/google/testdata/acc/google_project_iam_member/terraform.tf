provider "google" {}

terraform {
  required_version = "~> 0.15.0"
  required_providers {
    google = {
      version = "3.78.0"
    }
  }
}

resource "google_project_iam_member" "elie1" {
  role   = "roles/editor"
  member = "user:cloud-context-team@snyk.io"
}

resource "google_project_iam_member" "will1" {
  role   = "roles/viewer"
  member = "user:cloud-context-team@snyk.io"
}
