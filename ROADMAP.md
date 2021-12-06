# driftctl Roadmap (S1 2021)

This roadmap does not describe all the work that will be included within this timeframe, but it does describe our focus. We will include other work as events occur.

## Summary

* A more complete AWS support
* Support at least one new provider
* Maintain focus on Terraform support (no new IaC provider integration, like Pulumi)

## Resources

* Improve existing support for VPC, Route53, Lambda, S3, EC2, RDS Aurora
* Add support for:
  * API Gateway v1 & v2
  * SNS, SQS
  * ECR, ECS, EKS
  * Cloudfront
  * KMS
  * DynamoDB

## Providers

* Add GitHub support (at least repositories, organizations, users)
* Add initial support for either Azure or GCP (TBD)

## Issues & Enhancements

* Migration to Go 1.16 to support Apple Silicon
* Acceptance tests automation
* Don't scan for ignored or filtered resources (performance improvement)

## Disclosures

The product-development initiatives in this document reflect Snyk's current plans and are subject to change and/or cancellation in Snyk's sole discretion.
