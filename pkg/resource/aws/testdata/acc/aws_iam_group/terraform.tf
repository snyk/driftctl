provider "aws" {
    region = "us-east-1"
}

terraform {
    required_providers {
        aws = "5.94.1"
    }
}

resource "aws_iam_group" "my_developers" {
    name = "developers"
    path = "/users/"
}
