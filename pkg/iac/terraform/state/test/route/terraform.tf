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

resource "aws_route_table" "r" {
  vpc_id = aws_vpc.vpc.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.main.id
  }

  route {
    ipv6_cidr_block = "::/0"
    gateway_id = aws_internet_gateway.main.id
  }

  tags = {
    Name = "r"
  }
}

resource "aws_route_table" "rr" {
  vpc_id = aws_vpc.vpc.id

  tags = {
    Name = "rr"
  }
}

resource "aws_route" "route1" {
  route_table_id = aws_route_table.rr.id
  gateway_id = aws_internet_gateway.main.id
  destination_cidr_block = "1.1.1.1/32"
}

resource "aws_route" "route_v6" {
  route_table_id = aws_route_table.rr.id
  gateway_id = aws_internet_gateway.main.id
  destination_ipv6_cidr_block = "::/0"
}
