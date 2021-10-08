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

resource "aws_vpc" "vpc" {
    cidr_block = "10.100.0.0/16"
}

resource "aws_subnet" "subnet" {
    vpc_id = aws_vpc.vpc.id
    cidr_block = "10.100.0.0/24"
}

resource "aws_lb" "example" {
    name               = "example-${random_string.prefix.result}"
    internal           = true
    load_balancer_type = "network"

    subnet_mapping {
        subnet_id = aws_subnet.subnet.id
    }
}

resource "aws_api_gateway_vpc_link" "foo" {
    name        = "foo"
    description = "Description"
    target_arns = [aws_lb.example.arn]
}
