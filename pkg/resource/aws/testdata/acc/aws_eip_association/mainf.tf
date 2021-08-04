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

resource "aws_instance" "web" {
  ami           = data.aws_ami.ubuntu.id
  instance_type = "t3.micro"
  subnet_id     = aws_subnet.subnet-1.id
  availability_zone = "us-east-1b"

  tags = {
    Name = "HelloWorld"
  }
}

resource "aws_eip" "lb" {
  instance                  = aws_instance.web.id
  vpc                       = true
//   associate_with_private_ip = "10.0.0.12"
  depends_on                = [aws_internet_gateway.gw]
}

resource "aws_vpc" "default" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
}

resource "aws_internet_gateway" "gw" {
  vpc_id = aws_vpc.default.id
}

resource "aws_subnet" "subnet-1" {
  vpc_id                  = aws_vpc.default.id
  cidr_block              = "10.0.0.0/24"
  map_public_ip_on_launch = true
  availability_zone = "us-east-1b"

  depends_on = [aws_internet_gateway.gw]
}

