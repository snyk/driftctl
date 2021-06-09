provider "aws" {
  region = "us-east-1"
}

locals {
    timestamp = formatdate("YYYYMMDDhhmmss", timestamp())
    prefix = "vpc-${local.timestamp}"
}

terraform {
    required_providers {
        aws = "3.19.0"
    }
}
resource "aws_default_vpc" "default" {
    tags = {
        Name = "Default VPC"
    }
}

resource "aws_vpc" "vpc1" {
    cidr_block = "10.10.0.0/16"
    tags = {
        Name: "${local.prefix}-vpc1"
    }
}

resource "aws_vpc" "vpc2" {
    cidr_block = "10.11.0.0/16"
    tags = {
        Name: "${local.prefix}-vpc2"
    }
}

resource "aws_vpc" "vpc3" {
    cidr_block = "10.12.0.0/16"
    tags = {
        Name: "${local.prefix}-vpc2"
    }
}
