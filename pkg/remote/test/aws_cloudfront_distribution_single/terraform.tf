provider "aws" {
  region = "us-east-1"
}

terraform {
  required_providers {
    aws = "3.19.0"
  }
}

resource "aws_s3_bucket" "foo_cloudfront" {
  bucket = "foo-cloudfront"
  acl    = "private"
}

locals {
  s3_origin_id = "S3-foo-cloudfront"
}

resource "aws_cloudfront_distribution" "foo_distribution" {
  enabled             = false

  origin {
    domain_name = aws_s3_bucket.foo_cloudfront.bucket_regional_domain_name
    origin_id   = local.s3_origin_id
  }

  default_cache_behavior {
    allowed_methods  = ["GET", "HEAD"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = local.s3_origin_id
    viewer_protocol_policy = "allow-all"

    forwarded_values {
      query_string = false

      cookies {
        forward = "none"
      }
    }
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }
}
