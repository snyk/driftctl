provider "aws" {
    region = "us-east-1"
}

terraform {
    required_providers {
        aws = "3.19.0"
    }
}

resource "random_string" "prefix" {
    length  = 6
    upper   = false
    special = false
}

resource "aws_s3_bucket" "bucket" {
    bucket = "${random_string.prefix.result}-driftctl-test"
}

resource "aws_s3_bucket_public_access_block" "block" {
    bucket = aws_s3_bucket.bucket.id
}

resource "aws_s3_bucket" "bucket2" {
    bucket = "${random_string.prefix.result}-driftctl-test-2"
}

resource "aws_s3_bucket_public_access_block" "block2" {
    bucket = aws_s3_bucket.bucket2.id
    block_public_acls = true
    block_public_policy = true
}
