provider "aws" {
  region = "us-east-1"
}

terraform {
  required_providers {
    aws = "3.19.0"
  }
}

resource "aws_vpc" "vpc" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_default_route_table" "default" {
  default_route_table_id = aws_vpc.vpc.default_route_table_id
}

resource "aws_route_table" "r" {
  vpc_id = aws_vpc.vpc.id

  tags = {
    Name = "r"
  }
}
