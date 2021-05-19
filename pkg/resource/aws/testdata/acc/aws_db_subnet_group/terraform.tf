provider "aws" {
  region = "us-east-1"
}

terraform {
  required_providers {
    aws = "3.19.0"
  }
}

resource "aws_vpc" "vpc" {
    cidr_block = "10.1.0.0/16"
}

resource "aws_subnet" "subnet1" {
    vpc_id = aws_vpc.vpc.id
    cidr_block = "10.1.0.0/24"
    availability_zone = "us-east-1a"
}

resource "aws_subnet" "subnet2" {
    vpc_id = aws_vpc.vpc.id
    cidr_block = "10.1.1.0/24"
    availability_zone = "us-east-1b"
}

resource "aws_subnet" "subnet3" {
    vpc_id = aws_vpc.vpc.id
    cidr_block = "10.1.2.0/24"
    availability_zone = "us-east-1c"
}

resource "aws_subnet" "subnet4" {
    vpc_id = aws_vpc.vpc.id
    cidr_block = "10.1.3.0/24"
    availability_zone = "us-east-1a"
}

resource "aws_db_subnet_group" "foo" {
    name       = "foo"
    subnet_ids = [aws_subnet.subnet1.id, aws_subnet.subnet2.id]
}

resource "aws_db_subnet_group" "bar" {
    name_prefix       = "bar"
    subnet_ids = [aws_subnet.subnet3.id, aws_subnet.subnet4.id]
}
