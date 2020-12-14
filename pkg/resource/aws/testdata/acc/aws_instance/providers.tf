provider "aws" {
  region  = "eu-west-3"
}
terraform {
  required_providers {
    aws = {
      version = "~> 3.19.0"
    }
  }
}