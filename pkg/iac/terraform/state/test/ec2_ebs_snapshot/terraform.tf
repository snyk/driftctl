provider "aws" {
  region = "us-east-1"
}

terraform {
  required_providers {
    aws = "3.19.0"
  }
}

resource "aws_ebs_volume" "test-ebs-volume" {
    availability_zone = "us-east-1a"
    size              = 10

    tags = {
        Name = "HelloWorld"
    }
}

resource "aws_ebs_snapshot" "test-ebs-snapshot" {
    volume_id = aws_ebs_volume.test-ebs-volume.id

    tags = {
        Name = "HelloWorld_snap"
    }
}
