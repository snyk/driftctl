# IaC source

Currently, driftctl only supports reading IaC from a Terraform state.
We are investigating to support the Terraform code as well, as a state does not represent an intention.

## Supported IaC sources

* Terraform state
  * Local: `--from tfstate://terraform.tfstate`
  * S3: `--from tfstate+s3://my-bucket/path/to/state.tfstate`

### S3

driftctl needs to read your state, you could use the policy below to ensure minimal access to your state file

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": "s3:ListBucket",
      "Resource": "arn:aws:s3:::mybucket"
    },
    {
      "Effect": "Allow",
      "Action": ["s3:GetObject", "s3:PutObject"],
      "Resource": "arn:aws:s3:::mybucket/path/to/my/key"
    }
  ]
}
```
