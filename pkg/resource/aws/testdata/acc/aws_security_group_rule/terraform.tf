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
    prefix = "sgrule-${local.timestamp}"
}

resource "aws_vpc" "vpc" {
  cidr_block = "10.7.0.0/16"
  tags = {
    Name = "${local.prefix}-vpc"
  }
}

resource "aws_vpc_endpoint" "s3" {
    vpc_id       = aws_vpc.vpc.id
    service_name = "com.amazonaws.us-east-1.s3"
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

resource "aws_security_group_rule" "tutu-egress" {
    cidr_blocks       = ["0.0.0.0/0"]
    type              = "egress"
    description       = "Bar Full Open"
    from_port         = 0
    to_port           = 0
    protocol          = "tcp"
    security_group_id = aws_security_group.sg-bis-tutu.id
}

resource "aws_security_group_rule" "bla-ingress" {
    type             = "ingress"
    description      = "Bla 1"
    from_port        = 0
    to_port          = 0
    protocol         = "-1"
    ipv6_cidr_blocks = ["::/0"]
    security_group_id = aws_security_group.sg-bis-tutu.id
}

resource "aws_security_group" "sg-bis-titi" {
    name = "titi"

    tags = {
        Name = "Titi SG"
    }
}

resource "aws_security_group_rule" "test-ingress-custom-icmp" {
    type              = "ingress"
    protocol          = "1"
    cidr_blocks       = ["0.0.0.0/0"]
    security_group_id = aws_security_group.sg-bis-titi.id
    from_port         = 8
    to_port           = -1
}

resource "aws_security_group_rule" "test-ingress-custom-all" {
    type              = "egress"
    protocol          = "all"
    ipv6_cidr_blocks  = ["::/0"]
    cidr_blocks       = ["0.0.0.0/0"]
    security_group_id = aws_security_group.sg-bis-titi.id
    from_port         = 123
    to_port           = 42
}

resource "aws_security_group_rule" "test-ingress-custom-tcp" {
    type              = "ingress"
    protocol          = "6"
    cidr_blocks       = ["0.0.0.0/0"]
    security_group_id = aws_security_group.sg-bis-titi.id
    from_port         = 6
    to_port           = 42
}

resource "aws_security_group_rule" "test-ingress-custom-udp" {
    type              = "ingress"
    protocol          = "17"
    cidr_blocks       = ["0.0.0.0/0"]
    security_group_id = aws_security_group.sg-bis-titi.id
    from_port         = 6
    to_port           = 42
}

resource "aws_security_group" "sg-bis-4" {
    name = "tata"

    tags = {
        Name = "TATA SG"
    }
}

resource "aws_security_group_rule" "test-ingress-icmp" {
    type              = "ingress"
    protocol          = "icmp"
    cidr_blocks       = ["0.0.0.0/0"]
    security_group_id = aws_security_group.sg-bis-4.id
    from_port         = 8
    to_port           = -1
}

resource "aws_security_group_rule" "test-ingress-icmpv6" {
    type              = "ingress"
    protocol          = "icmpv6"
    cidr_blocks       = ["0.0.0.0/0"]
    security_group_id = aws_security_group.sg-bis-4.id
    from_port         = -1
    to_port           = -1
}

resource "aws_security_group_rule" "test-ingress-bgp" {
    type              = "ingress"
    protocol          = "3"
    cidr_blocks       = ["0.0.0.0/0"]
    security_group_id = aws_security_group.sg-bis-4.id
    from_port         = 10
    to_port           = 55
}

resource "aws_security_group" "sg-bis-third" {
    name = "baz"

    tags = {
        Name = "Baz SG"
    }
}

resource "aws_security_group_rule" "baz-egress" {
    cidr_blocks       = ["0.0.0.0/0"]
    ipv6_cidr_blocks  = ["::/0"]
    type              = "egress"
    description       = "Bar Full Open"
    from_port         = 0
    to_port           = 0
    protocol          = "-1"
    security_group_id = aws_security_group.sg-bis-third.id
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

resource "aws_security_group_rule" "bar-egress" {
    cidr_blocks       = ["0.0.0.0/0"]
    ipv6_cidr_blocks  = ["::/0"]
    type              = "egress"
    description       = "Bar Full Open"
    from_port         = 0
    to_port           = 0
    protocol          = "-1"
    security_group_id = aws_security_group.sg-bis.id
}

resource "aws_security_group_rule" "bar-egress-stream" {
    cidr_blocks       = ["0.0.0.0/0"]
    type              = "egress"
    description       = "Stream"
    from_port         = 10
    to_port           = 55
    protocol          = "5"
    security_group_id = aws_security_group.sg-bis.id
}

resource "aws_security_group" "test-sg" {
    name = "foo"

    tags = {
        Name = "Foo SG"
    }
}

resource "aws_security_group_rule" "test-ingress-rule" {
    type              = "ingress"
    from_port         = 0
    to_port           = 65535
    protocol          = "tcp"
    security_group_id = aws_security_group.test-sg.id
    self              = true
    description       = "Test 1"
}

resource "aws_security_group_rule" "test-ingress-rule-bis" {
    type              = "ingress"
    from_port         = 0
    to_port           = 0
    protocol          = "icmp"
    security_group_id = aws_security_group.test-sg.id
    self              = true
}

resource "aws_security_group_rule" "ingress" {
    cidr_blocks       = ["0.0.0.0/0"]
    type              = "ingress"
    description       = "Foo 1"
    from_port         = 0
    to_port           = 0
    protocol          = "-1"
    security_group_id = aws_security_group.test-sg.id
}

resource "aws_security_group_rule" "foo" {
    cidr_blocks       = ["1.2.0.0/16", "5.6.7.0/24"]
    type              = "ingress"
    description       = "Foo 5"
    from_port         = 0
    to_port           = 0
    protocol          = "-1"
    security_group_id = aws_security_group.test-sg.id
}

resource "aws_security_group_rule" "baz-ingress" {
    type              = "ingress"
    description       = "Baz 2"
    from_port         = 0
    to_port           = 0
    protocol          = "tcp"
    security_group_id = aws_security_group.test-sg.id
    prefix_list_ids   = [aws_vpc_endpoint.s3.prefix_list_id]
}

resource "aws_security_group_rule" "egress" {
    cidr_blocks       = ["0.0.0.0/0"]
    ipv6_cidr_blocks  = ["::/0"]
    type              = "egress"
    description       = "Bar 1"
    from_port         = 0
    to_port           = 0
    protocol          = "-1"
    security_group_id = aws_security_group.test-sg.id
}
