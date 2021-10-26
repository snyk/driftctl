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

resource "google_sql_database_instance" "instance" {
    name   = "dctl-qa-${random_string.postfix.result}"
    region = "us-central1"
    settings {
        tier = "db-f1-micro"
    }

    deletion_protection  = "false"
}
