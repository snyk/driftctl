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
    prefix = "dbsubnet-${local.timestamp}"
}

resource "aws_vpc" "vpc" {
    cidr_block = "10.1.0.0/16"
    tags = {
        Name: "${local.prefix}-vpc"
    }
}

resource "aws_subnet" "subnet1" {
    vpc_id = aws_vpc.vpc.id
    cidr_block = "10.1.0.0/24"
    availability_zone = "us-east-1a"
    tags = {
        Name: "${local.prefix}-subnet1"
    }
}

resource "aws_subnet" "subnet2" {
    vpc_id = aws_vpc.vpc.id
    cidr_block = "10.1.1.0/24"
    availability_zone = "us-east-1b"
    tags = {
        Name: "${local.prefix}-subnet2"
    }
}

resource "aws_subnet" "subnet3" {
    vpc_id = aws_vpc.vpc.id
    cidr_block = "10.1.2.0/24"
    availability_zone = "us-east-1c"
    tags = {
        Name: "${local.prefix}-subnet3"
    }
}

resource "aws_subnet" "subnet4" {
    vpc_id = aws_vpc.vpc.id
    cidr_block = "10.1.3.0/24"
    availability_zone = "us-east-1a"
    tags = {
        Name: "${local.prefix}-subnet4"
    }
}

resource "aws_db_subnet_group" "foo" {
    name       = "foo"
    subnet_ids = [aws_subnet.subnet1.id, aws_subnet.subnet2.id]
}

resource "aws_db_subnet_group" "bar" {
    name_prefix       = "bar"
    subnet_ids = [aws_subnet.subnet3.id, aws_subnet.subnet4.id]
}
