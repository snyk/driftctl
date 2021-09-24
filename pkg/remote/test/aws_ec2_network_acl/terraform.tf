provider "aws" {
    region = "us-east-1"
}

terraform {
    required_providers {
        aws = {
            version = "3.47.0"
        }
    }
}

resource "aws_default_vpc" "default" {}

resource "aws_network_acl" "test" {
    vpc_id = aws_default_vpc.default.id
    tags = {
        Name = "test2"
    }
}

resource "aws_network_acl_rule" "ingress" {
    network_acl_id = aws_network_acl.test.id
    protocol       = "tcp"
    rule_action    = "allow"
    rule_number    = 100
    from_port = 80
    to_port = 80
    cidr_block = aws_default_vpc.default.cidr_block
}

resource "aws_network_acl_rule" "egress" {
    network_acl_id = aws_network_acl.test.id
    egress = true
    protocol       = "udp"
    rule_action    = "allow"
    rule_number    = 100
    from_port = 80
    to_port = 80
    cidr_block = aws_default_vpc.default.cidr_block
}

resource "aws_network_acl" "test2" {
    vpc_id = aws_default_vpc.default.id

    egress = [
        {
            protocol   = "udp"
            rule_no    = 100
            action     = "allow"
            cidr_block = aws_default_vpc.default.cidr_block
            from_port  = 80
            to_port    = 80
            icmp_code = 0
            icmp_type = 0
            ipv6_cidr_block = ""
        }
    ]


    ingress = [
        {
            protocol   = "udp"
            rule_no    = 100
            action     = "allow"
            cidr_block = aws_default_vpc.default.cidr_block
            from_port  = 80
            to_port    = 80
            icmp_code = 0
            icmp_type = 0
            ipv6_cidr_block = ""
        }
    ]

    tags = {
        Name = "test"
    }
}


