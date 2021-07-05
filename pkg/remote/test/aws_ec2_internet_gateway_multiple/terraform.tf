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

resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.vpc.id
  tags = {
    Name = "main"
  }
}

resource "aws_vpc" "vpc_bis" {
  cidr_block = "10.1.0.0/16"
}

resource "aws_internet_gateway" "foo" {
  vpc_id = aws_vpc.vpc_bis.id
  tags = {
    Name = "foo"
  }
}
