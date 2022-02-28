provider "aws" {
  region  = "us-east-1"
}

terraform {
  required_providers {
    aws = {
      version = "3.19.0"
    }
  }

  backend "s3" {
    bucket = "driftctl-acc-statereader-multiples-states"
    key    = "states/s3/state1"
    region = "us-east-1"
  }
}

resource "random_string" "prefix" {
  length  = 6
  upper   = false
  special = false
}

resource "aws_s3_bucket" "foobar" {
  bucket = "${random_string.prefix.result}.driftctl-test.com"
}
