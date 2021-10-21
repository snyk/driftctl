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
    dataset_id                  = "TestAcc_Google_BigqueryTable"
    friendly_name               = "TestAcc_Google_BigqueryTable"
    description                 = "This is a test description"
    location                    = "EU"
}

resource "google_bigquery_table" "default" {
    dataset_id = google_bigquery_dataset.dataset.dataset_id
    table_id   = "bar"
    deletion_protection = false

    schema = <<EOF
[
  {
    "name": "bar",
    "type": "STRING",
    "mode": "NULLABLE",
    "description": "foobar"
  }
]
EOF

}
