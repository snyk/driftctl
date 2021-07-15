provider "aws" {
  region = "us-east-1"
}
terraform {
  required_providers {
    aws = {
      version = "3.19.0"
    }
  }
}

resource "aws_iam_user" "testuser_access_key" {
    name = "testuser_access_key"
}

resource "aws_iam_access_key" "accesskey" {
    user = aws_iam_user.testuser_access_key.name
}

resource "aws_iam_access_key" "ak" {
    user    = aws_iam_user.test.name
    pgp_key = "keybase:test"
}

resource "aws_iam_user" "test" {
    name = "test"
    path = "/test/"
}
