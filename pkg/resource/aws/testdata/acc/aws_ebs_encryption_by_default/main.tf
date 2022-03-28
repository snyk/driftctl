provider "aws" {
  region = "us-east-1"
}

resource "aws_ebs_encryption_by_default" "test-encryption" {
    enabled = true
}
