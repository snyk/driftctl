provider "aws" {
  region = "eu-west-3"
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
  cidr_block = "10.0.0.0/16"
}

resource "aws_vpc" "vpc2" {
  cidr_block = "10.1.0.0/16"
}

resource "aws_vpc" "vpc3" {
  cidr_block = "10.2.0.0/16"
}
