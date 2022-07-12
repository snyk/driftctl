terraform {
    backend "gcs" {
        bucket  = "tf-state-prod"
        prefix  = "terraform/state"
    }
}

provider "google" {
    project = "my-project"
    region  = "us-central1"
    zone    = "us-central1-c"
}
