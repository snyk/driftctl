provider "aws" {
  region = "us-east-1"
}

terraform {
  required_providers {
    aws = "3.19.0"
  }
}

locals {
    timestamp = formatdate("YYYYMMDDhhmmss", timestamp())
    prefix = "igw-${local.timestamp}"
}

resource "aws_vpc" "vpc" {
  cidr_block = "10.2.0.0/16"
  tags = {
    Name: "${local.prefix}-default"
  }
}

resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.vpc.id

  tags = {
    Name = "main"
  }
}
