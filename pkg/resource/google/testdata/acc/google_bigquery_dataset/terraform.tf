provider "google" {}

terraform {
    required_version = "~> 0.15.0"
    required_providers {
        google = {
            version = "3.78.0"
        }
    }
}

resource "google_bigquery_dataset" "dataset" {
    dataset_id                  = "TestAcc_Google_BigqueryDataset"
    friendly_name               = "TestAcc_Google_BigqueryDataset"
    description                 = "This is a test description"
    location                    = "EU"
}
