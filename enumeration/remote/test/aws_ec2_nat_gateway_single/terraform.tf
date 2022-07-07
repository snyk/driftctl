provider "aws" {
  region = "us-east-1"
}

terraform {
  required_providers {
    aws = "3.19.0"
  }
}

resource "aws_default_subnet" "default" {
  availability_zone = "us-east-1a"
}

resource "aws_eip" "default" {}

resource "aws_nat_gateway" "nat1" {
  allocation_id = aws_eip.default.id
  subnet_id = aws_default_subnet.default.id
  tags = {
    Name = "nat1"
  }
}
