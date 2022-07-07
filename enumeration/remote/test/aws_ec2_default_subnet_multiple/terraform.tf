provider "aws" {
  region = "us-east-1"
}

terraform {
  required_providers {
    aws = "3.19.0"
  }
}

resource "aws_vpc" "vpc_for_subnets" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_default_subnet" "default-a" {
  availability_zone = "us-east-1a"
}

resource "aws_default_subnet" "default-b" {
  availability_zone = "us-east-1b"
}

resource "aws_default_subnet" "default-c" {
  availability_zone = "us-east-1c"
}

resource "aws_subnet" "subnet1" {
  vpc_id = aws_vpc.vpc_for_subnets.id
  cidr_block = "10.0.0.0/24"
}

resource "aws_subnet" "subnet2" {
  vpc_id = aws_vpc.vpc_for_subnets.id
  cidr_block = "10.0.1.0/24"
}

resource "aws_subnet" "subnet3" {
  vpc_id = aws_vpc.vpc_for_subnets.id
  cidr_block = "10.0.2.0/24"
}
