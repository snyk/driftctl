provider "aws" {
  version = "3.5.0"
  region  = "eu-west-3"
}

resource "aws_route53_zone" "foobar" {
  name  = "foo-${count.index}.com"
  count = 3
}

output "foo-0" {
  value = aws_route53_zone.foobar[0].zone_id
}

output "foo-1" {
  value = aws_route53_zone.foobar[1].zone_id
}

output "foo-2" {
  value = aws_route53_zone.foobar[2].zone_id
}