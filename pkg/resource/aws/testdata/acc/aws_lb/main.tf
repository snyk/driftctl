provider "aws" {
  region = "us-east-1"
}

resource "aws_vpc" "main" {
    cidr_block = "10.100.0.0/16"
}

resource "aws_internet_gateway" "gw" {
    vpc_id = aws_vpc.main.id
}

resource "aws_subnet" "main-1" {
    vpc_id            = aws_vpc.main.id
    cidr_block        = "10.100.0.0/24"
    availability_zone = "us-east-1a"
}

resource "aws_subnet" "main-2" {
    vpc_id            = aws_vpc.main.id
    cidr_block        = "10.100.1.0/24"
    availability_zone = "us-east-1b"
}

resource "aws_security_group" "lb_sg" {
    name        = "acc_allow_tls_lb"
    description = "Allow TLS inbound traffic"
    vpc_id      = aws_vpc.main.id

    ingress {
        description = "TLS from VPC"
        from_port   = 443
        to_port     = 443
        protocol    = "tcp"
        cidr_blocks = [aws_vpc.main.cidr_block]
    }

    egress {
        from_port        = 0
        to_port          = 0
        protocol         = "-1"
        cidr_blocks      = ["0.0.0.0/0"]
        ipv6_cidr_blocks = ["::/0"]
    }
}

resource "aws_lb" "test" {
    name               = "acc-test-lb-tf"
    internal           = false
    load_balancer_type = "application"
    security_groups    = [aws_security_group.lb_sg.id]
    subnets            = [aws_subnet.main-1.id,aws_subnet.main-2.id]
    enable_deletion_protection = false
}
