# Known Issues and Limitations

## AWS Regions & Credentials Limits

- The user needs to use the same AWS region and credentials for both the scanned infrastructure and the S3 bucket where the Terraform state is stored (for example, a Terraform state stored on S3 on us-east-1 for an infrastructure to be scanned on us-west-1 won't work). Think `AWS_PROFILE` for the underlying reason. See the related [GitHub Discussion](https://github.com/cloudskiff/driftctl/discussions/130).
- Driftctl currently doesn't support multiple aliased providers in a single Terraform state (like a single account but multiple regions). This will be implemented soon.


## Terraform & Providers Support

- Terraform version >= 0.12 is supported
- Terraform AWS provider version >= 3.x is supported

## Terraform Resources

### AWS

- aws_security_group and aws_security_group_rule:

For security group that has in-line egress or ingress rules, driftctl will output an alert message at the end of the scan to warn you that those rules are falsely unmanaged. The explanation is that we can't distinct, based only on the Terraform state, rules created in the console and rules created in-line in either egress or ingress blocks.

### Github
- github_branch_protection_v3:
  - This resource is not supported and will probably never be as it overlaps with github_branch_protection.
    `github_branch_protection` is more suitable for performance purpose.
    We cannot support these two resources as we don't have any way to discriminate them by enumerating resources from
    remote side. They represent the same notion but are from two different APIs (REST vs GraphQL).
    driftctl team recommends you to use the newer `github_branch_protection`, or at least ignore all your `github_branch_protection_v3` in driftignore.
- github_branch_protection:
  - Branch protection resources are not returned as unmanaged if the branch protection pattern does not match at least one branch.
  - We cannot show the related repository name in driftctl output as the terraform provider does not retrieve this information.
