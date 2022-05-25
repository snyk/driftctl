provider "aws" {
    region = "us-east-1"
}

terraform {
    required_providers {
        aws = "3.75.1"
    }
}

resource "aws_vpc" "example" {
    cidr_block                       = "10.1.0.0/16"
}

resource "aws_ec2_managed_prefix_list" "example" {
    name           = "example"
    address_family = "IPv4"
    max_entries    = 5
}

resource "aws_route_table" "example" {
    vpc_id = aws_vpc.example.id
}

resource "aws_subnet" "example" {
    vpc_id     = aws_vpc.example.id
    cidr_block = "10.1.1.0/24"
}

resource "aws_nat_gateway" "example" {
    connectivity_type = "private"
    subnet_id         = aws_subnet.example.id
}

resource "aws_route" "r" {
    route_table_id              = aws_route_table.example.id
    nat_gateway_id = aws_nat_gateway.example.id
    destination_prefix_list_id = aws_ec2_managed_prefix_list.example.id
}
