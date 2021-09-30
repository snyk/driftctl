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

resource "aws_default_network_acl" "default" {
    default_network_acl_id = aws_default_vpc.default.default_network_acl_id

    ingress {
        protocol   = "tcp"
        rule_no    = 100
        action     = "allow"
        cidr_block = "0.0.0.0/0"
        from_port  = 0
        to_port    = 0
    }

    egress {
        protocol   = "udp"
        rule_no    = 100
        action     = "allow"
        cidr_block = "0.0.0.0/0"
        from_port  = 0
        to_port    = 0
    }

    lifecycle {
        ignore_changes = [subnet_ids]
    }
}
