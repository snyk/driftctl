provider "aws" {
  region = "us-east-1"
}

terraform {
  required_providers {
    aws = "3.19.0"
  }
}

resource "aws_api_gateway_api_key" "foo" {
    name = "foo"
    description = "Foo Api Key"
    enabled = false
}

resource "aws_api_gateway_api_key" "bar" {
    name = "bar"
    description = "Bar Api Key"
    enabled = false
}
