provider "aws" {
  region = "us-east-1"
}

terraform {
  required_providers {
    aws = "3.62.0"
  }
}

resource "aws_default_vpc" "default" {}

resource "aws_vpc" "vpc" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_default_route_table" "default" {
  default_route_table_id = aws_default_vpc.default.default_route_table_id
}

resource "aws_route_table" "rr" {
  vpc_id = aws_vpc.vpc.id

  tags = {
    Name = "rr"
  }

  timeouts {
    create = "6m"
    update = "3m"
    delete = "6m"
  }
}
