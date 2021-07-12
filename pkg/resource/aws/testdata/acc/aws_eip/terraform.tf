provider "aws" {
  region = "us-east-1"
}
terraform {
  required_providers {
    aws = {
      version = "3.44.0"
    }
  }
}

locals {
    timestamp = formatdate("YYYYMMDDhhmmss", timestamp())
    prefix = "eip-${local.timestamp}"
}

# data source for an official Ubuntu 20.04 AMI
data "aws_ami" "ubuntu" {
    most_recent = true

    filter {
        name   = "name"
        values = ["ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-server-*"]
    }

    filter {
        name   = "virtualization-type"
        values = ["hvm"]
    }

    owners = ["099720109477"] # Canonical
}

resource "aws_vpc" "default" {
    cidr_block           = "10.4.0.0/24"
    tags = {
        Name: "${local.prefix}-default"
    }
}

resource "aws_internet_gateway" "gw" {
    vpc_id = aws_vpc.default.id
}

resource "aws_eip_association" "eip_assoc" {
    instance_id   = aws_instance.instance.id
    allocation_id = aws_eip.example.id
}

resource "aws_subnet" "tf_test_subnet" {
    vpc_id                  = aws_vpc.default.id
    cidr_block              = "10.4.0.0/24"
    depends_on = [aws_internet_gateway.gw]
    availability_zone = "us-east-1a"
    tags = {
        Name: "${local.prefix}-tf_test_subnet"
    }
}

resource "aws_instance" "instance" {
    ami           = data.aws_ami.ubuntu.id
    availability_zone = "us-east-1a"
    instance_type     = "t3.micro"
    private_ip = "10.4.0.12"
    subnet_id  = aws_subnet.tf_test_subnet.id
    depends_on = [aws_internet_gateway.gw]
}

resource "aws_eip" "example" {
    vpc = true

    instance                  = aws_instance.instance.id
    depends_on                = [aws_internet_gateway.gw]
}
