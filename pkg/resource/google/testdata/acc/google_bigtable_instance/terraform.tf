provider "google" {}

terraform {
    required_version = "~> 0.15.0"
    required_providers {
        google = {
            version = "3.78.0"
        }
    }
}

resource "google_bigtable_instance" "test-instance" {
    name = "tf-instance"
    deletion_protection = false

    cluster {
        zone = "us-central1-a"
        cluster_id   = "tf-instance-cluster"
        num_nodes    = 1
        storage_type = "HDD"
    }
}
