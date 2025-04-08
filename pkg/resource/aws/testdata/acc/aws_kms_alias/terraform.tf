provider "aws" {
  region = "us-east-1"
}

terraform {
  required_providers {
    aws = "5.94.1"
  }
}

resource "aws_kms_key" "key" {
  deletion_window_in_days = 7
}

resource "aws_kms_alias" "foo" {
  name          = "alias/foo"
  target_key_id = aws_kms_key.key.key_id
}

resource "aws_kms_alias" "baz" {
  name_prefix   = "alias/baz"
  target_key_id = aws_kms_key.key.key_id
}
