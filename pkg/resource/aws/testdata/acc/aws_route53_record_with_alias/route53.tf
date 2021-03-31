resource "aws_route53_zone" "foo-zone" {
  name = "foo-2.com"
}

resource "aws_route53_record" "alias" {
    zone_id = aws_route53_zone.foo-zone.zone_id
    name    = "alias.foo-2.com"
    type    = "A"
    alias {
        evaluate_target_health = false
        name                   = aws_route53_record.www.name
        zone_id                = aws_route53_zone.foo-zone.zone_id
    }
}

resource "aws_route53_record" "www" {
    zone_id = aws_route53_zone.foo-zone.zone_id
    name    = "www.foo-2.com"
    type    = "A"
    ttl     = "300"
    records = ["1.1.1.1"]
}
