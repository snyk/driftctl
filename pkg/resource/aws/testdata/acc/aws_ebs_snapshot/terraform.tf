provider "aws" {
  region = "us-east-1"
}

terraform {
  required_providers {
    aws = "3.19.0"
  }
}

resource "aws_ebs_volume" "foo" {
    availability_zone = "us-east-1a"
    size              = 10

    tags = {
        Name = "Foo Volume"
    }
}

resource "aws_ebs_snapshot" "foo" {
    volume_id = aws_ebs_volume.foo.id

    tags = {
        Name = "Foo Snapshot"
    }

    timeouts {
        create = "20m"
    }
}
