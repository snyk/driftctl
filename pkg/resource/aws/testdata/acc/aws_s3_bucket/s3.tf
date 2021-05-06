resource "random_string" "prefix" {
  length  = 6
  upper   = false
  special = false
}

resource "aws_s3_bucket" "acl" {
    bucket = "${random_string.prefix.result}.acl.driftctl-test.com"
    acl = "public-read"
}

resource "aws_s3_bucket" "foobar" {
    bucket = "${random_string.prefix.result}.driftctl-test.com"
}

resource "aws_s3_bucket" "foobar-policy" {
    bucket = "${random_string.prefix.result}.policy.driftctl-test.com"
    policy = <<POLICY
{
  "Version":"2012-10-17",
  "Statement":[
    {
      "Sid":"PublicRead",
      "Effect":"Allow",
      "Principal": "*",
      "Action":["s3:GetObject","s3:GetObjectVersion"],
      "Resource":["arn:aws:s3:::${random_string.prefix.result}.policy.driftctl-test.com/*"]
    }
  ]
}
POLICY
}

resource "aws_s3_bucket_policy" "foobar" {
    bucket = aws_s3_bucket.foobar.bucket
    policy = <<POLICY
{
  "Version":"2012-10-17",
  "Statement":[
    {
      "Sid":"PublicRead",
      "Effect":"Allow",
      "Principal": "*",
      "Action":["s3:GetObject","s3:GetObjectVersion"],
      "Resource":["arn:aws:s3:::${random_string.prefix.result}.driftctl-test.com/*"]
    }
  ]
}
POLICY
}
