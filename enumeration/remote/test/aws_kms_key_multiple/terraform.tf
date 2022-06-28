provider "aws" {
  region = "us-east-1"
}

terraform {
  required_providers {
    aws = "3.19.0"
  }
}

resource "aws_kms_key" "foo" {
  description              = "Foo"
  deletion_window_in_days  = 10
  customer_master_key_spec = "RSA_4096"
}

resource "aws_kms_key" "bar" {
  description              = "Bar"
  customer_master_key_spec = "RSA_2048"
  key_usage = "SIGN_VERIFY"
}

resource "aws_kms_key" "baz" {
  description             = "Baz"
  deletion_window_in_days = 10
  is_enabled              = false
  tags = {
    "Foo" = "true"
  }
}
