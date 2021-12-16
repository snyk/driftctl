provider "aws" {
  version = "3.5.0"
  region  = "eu-west-3"
}
resource "aws_route53_zone" "foobar" {
  name    = "foo.bar"
  comment = "test comment"
  tags = {
    test = "example"
  }
}

output "zone_id" {
  value = aws_route53_zone.foobar.zone_id
}