provider "aws" {
  region  = "us-east-1"
}

terraform {
  required_providers {
    aws = {
      version = "5.94.1"
    }
  }

  backend "local" {
    path = "../states/route53/terraform.tfstate"
  }
}

resource "random_string" "prefix" {
  length  = 6
  upper   = false
  special = false
}

resource "aws_route53_zone" "foobar" {
  name = "${random_string.prefix.result}-example.com"
}
