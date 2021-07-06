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

resource "aws_default_route_table" "default" {
  default_route_table_id = aws_vpc.vpc.default_route_table_id

  tags = {
    Name = "default_table"
  }
}

resource "aws_route_table" "table2" {
  vpc_id = aws_vpc.vpc.id

  tags = {
    Name = "table2"
  }
}

resource "aws_route_table" "table1" {
  vpc_id = aws_vpc.vpc.id

  tags = {
    Name = "table1"
  }
}

resource "aws_route_table" "table3" {
  vpc_id = aws_vpc.vpc.id

  tags = {
    Name = "table3"
  }
}
