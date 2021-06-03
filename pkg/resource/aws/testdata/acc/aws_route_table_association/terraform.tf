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
    prefix = "rtbassoc-${local.timestamp}"
}

resource "aws_vpc" "vpc" {
  cidr_block = "10.5.0.0/16"
  tags = {
    Name = "${local.prefix}-vpc"
  }
}

resource "aws_route_table" "route" {
  vpc_id = aws_vpc.vpc.id
  tags = {
    Name = "route"
  }
}

resource "aws_route_table" "route2" {
  vpc_id = aws_vpc.vpc.id
  tags = {
    Name = "route2"
  }
}

resource "aws_subnet" "subnet" {
  cidr_block = "10.5.0.0/24"
  vpc_id = aws_vpc.vpc.id
  tags = {
    Name = "subnet"
  }
}

resource "aws_subnet" "subnet1" {
  cidr_block = "10.5.1.0/24"
  vpc_id = aws_vpc.vpc.id
  tags = {
    Name = "subnet1"
  }
}

resource "aws_subnet" "subnet2" {
  cidr_block = "10.5.2.0/24"
  vpc_id = aws_vpc.vpc.id
  tags = {
    Name = "subnet2"
  }
}

resource "aws_route_table_association" "assoc_route_subnet" {
  route_table_id = aws_route_table.route.id
  subnet_id = aws_subnet.subnet.id
}

resource "aws_route_table_association" "assoc_route_subnet1" {
  route_table_id = aws_route_table.route.id
  subnet_id = aws_subnet.subnet1.id
}

resource "aws_route_table_association" "assoc_route_subnet2" {
  route_table_id = aws_route_table.route.id
  subnet_id = aws_subnet.subnet2.id
}

resource "aws_internet_gateway" "gateway" {
  vpc_id = aws_vpc.vpc.id
  tags = {
    Name = "gateway"
  }
}

resource "aws_route_table_association" "assoc_route2_gateway" {
  route_table_id = aws_route_table.route2.id
  gateway_id = aws_internet_gateway.gateway.id
}
