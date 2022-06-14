provider "aws" {
  region = "us-east-1"
}

terraform {
  required_providers {
    aws = "3.62.0"
  }
}

locals {
    timestamp = formatdate("YYYYMMDDhhmmss", timestamp())
    prefix = "rtb-${local.timestamp}"
}

resource "aws_vpc" "vpc" {
  cidr_block = "10.1.0.0/16"
  tags = {
    Name: "${local.prefix}-default"
  }
}

resource "aws_default_route_table" "default" {
  default_route_table_id = aws_vpc.vpc.default_route_table_id
}

resource "aws_route_table" "r" {
  vpc_id = aws_vpc.vpc.id

  tags = {
    Name = "r"
  }

  timeouts {
    create = "6m"
    update = "3m"
    delete = "6m"
  }
}
