provider "aws" {
  region = "us-east-1"
}

terraform {
  required_providers {
    aws = "5.94.1"
  }
}

resource "aws_db_subnet_group" "foo" {
    name       = "foo"
    subnet_ids = ["subnet-23222e4a", "subnet-fdfdda86"]
}

resource "aws_db_subnet_group" "bar" {
    name_prefix       = "bar"
    subnet_ids = ["subnet-23222e4a", "subnet-fdfdda86"]
}
