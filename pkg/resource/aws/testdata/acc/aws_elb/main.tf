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

    tags = {
        Name = "HelloWorld"
    }
}

# Create a new load balancer
resource "aws_elb" "bar" {
    name               = "acc-test-terraform-elb"
    availability_zones = ["us-east-1a", "us-east-1b"]

    listener {
        instance_port     = 8000
        instance_protocol = "http"
        lb_port           = 80
        lb_protocol       = "http"
    }

    instances                   = [aws_instance.web.id]
    cross_zone_load_balancing   = true
    idle_timeout                = 400
    connection_draining         = true
    connection_draining_timeout = 400
}
