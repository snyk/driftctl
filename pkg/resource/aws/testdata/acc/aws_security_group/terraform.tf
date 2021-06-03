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
    prefix = "sg-${local.timestamp}"
}

resource "aws_vpc" "vpc" {
  cidr_block = "10.6.0.0/16"
  tags = {
    Name = "${local.prefix}-vpc"
  }
}

resource "aws_default_security_group" "default" {
  vpc_id = aws_vpc.vpc.id

  tags = {
    Name = "Default SG"
  }
}

resource "aws_security_group" "best-security-group-ever" {
    name = "best-security-group-ever"

    egress {
        from_port   = 0
        to_port     = 0
        protocol    = "-1"
        cidr_blocks = ["0.0.0.0/0"]
    }

    tags = {
        Name = "This is the best"
    }
}

resource "aws_security_group" "infra" {
    name        = "infra"
    description = "infra SSH"

    egress {
        from_port   = 0
        to_port     = 0
        protocol    = "-1"
        cidr_blocks = ["0.0.0.0/0"]
    }

    tags = {
        Name = "infra"
    }
}

resource "aws_security_group" "sg-bis-tutu-twice" {
    name = "tutu-twice"

    tags = {
        Name = "Tutu TWICE SG"
    }
}

resource "aws_security_group" "sg-bis-tutu" {
    name = "tutu"

    tags = {
        Name = "Tutu SG"
    }
}

resource "aws_security_group" "sg-bis-titi" {
    name = "titi"

    tags = {
        Name = "Titi SG"
    }
}

resource "aws_security_group" "sg-bis-4" {
    name = "tata"

    tags = {
        Name = "TATA SG"
    }
}

resource "aws_security_group" "sg-bis-third" {
    name = "baz"

    tags = {
        Name = "Baz SG"
    }
}

resource "aws_security_group" "sg-bis" {
    name = "bar"

    ingress {
        description = "TLS from VPC"
        from_port   = 443
        to_port     = 443
        protocol    = "tcp"
        cidr_blocks = ["0.0.0.0/0"]
    }
    tags = {
        Name = "Bar SG"
    }
}

resource "aws_security_group" "test-sg" {
    name = "foo"

    tags = {
        Name = "Foo SG"
    }
}
