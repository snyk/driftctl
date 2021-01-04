resource "aws_route53_zone" "foo-zone" {
  name = "foo-2.com"
}

resource "aws_route53_record" "foo-record" {
  zone_id = aws_route53_zone.foo-zone.zone_id
  name    = "test0"
  type    = "TXT"
  ttl     = 300
  records = ["test0"]
}

resource "aws_route53_record" "foo-record-2" {
  zone_id = aws_route53_zone.foo-zone.zone_id
  name    = "test1.foo-2.com"
  type    = "TXT"
  ttl     = 300
  records = ["test1.foo-2.com"]
}

resource "aws_route53_record" "foo-record-bis-3" {
  zone_id = aws_route53_zone.foo-zone.zone_id
  name    = "_test2.foo-2.com"
  type    = "TXT"
  ttl     = 300
  records = ["_test2.foo-2.com"]
}

resource "aws_route53_record" "foo-record-bis-4" {
  zone_id = aws_route53_zone.foo-zone.zone_id
  name    = "test3."
  type    = "TXT"
  ttl     = 300
  records = ["test3."]
}
