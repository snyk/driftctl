provider "google" {}

terraform {
    required_version = "~> 0.15.0"
    required_providers {
        google = {
            version = "3.78.0"
        }
    }
}

resource "google_bigtable_instance" "instance" {
    name = "tf-instance-table"
    deletion_protection = false

    cluster {
        zone = "us-central1-a"
        cluster_id   = "tf-instance-cluster-table"
        num_nodes    = 1
        storage_type = "HDD"
    }
}

resource "google_bigtable_table" "table" {
    name          = "tf-table"
    instance_name = google_bigtable_instance.instance.name
}
