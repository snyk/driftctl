provider "aws" {
    region = "us-east-1"
}

terraform {
    required_providers {
        aws = "3.19.0"
    }
}

resource "aws_s3_account_public_access_block" "example" {
    block_public_acls   = true
    block_public_policy = true
}

