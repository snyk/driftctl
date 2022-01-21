provider "aws" {
  region = "us-east-1"
}

terraform {
  required_providers {
    aws = "3.19.0"
  }
}

resource "aws_vpc" "vpc" {
    cidr_block = "10.100.0.0/16"
}

resource "aws_subnet" "subnet" {
    vpc_id = aws_vpc.vpc.id
    cidr_block = "10.100.0.0/24"
    availability_zone_id = "use1-az1"
}

resource "aws_security_group" "foo" {
    vpc_id = aws_vpc.vpc.id
}

resource "aws_apigatewayv2_vpc_link" "foo" {
    name               = "foo"
    security_group_ids = [aws_security_group.foo.id]
    subnet_ids         = [aws_subnet.subnet.id]
    tags = {
        Usage = "example"
    }
}
