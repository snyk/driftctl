provider "aws" {
  region  = "us-east-1"
}

terraform {
  required_providers {
    aws = {
      version = "3.19.0"
    }
  }

  backend "local" {
    path = "../states/s3/terraform.tfstate"
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
