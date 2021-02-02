# Known Issues and Limitations

## AWS Regions Limits

- The user needs to use the same AWS region for both the scanned infrastructure and the S3 bucket where the Terraform state is stored (for example, a Terraform state stored on S3 on us-east-1 for an infrastructure to be scanned on us-west-1 won't work.).
  - See the related [GitHub Discussion](https://github.com/cloudskiff/driftctl/discussions/130).
