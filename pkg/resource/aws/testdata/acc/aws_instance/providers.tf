terraform {
  required_version = "= 0.12.29"
  required_providers {
    aws    = "= 3.5.0"
    random = "= 3.0.0"
  }
}

provider "aws" {
  version = "3.5.0"
  region  = "eu-west-3"
}