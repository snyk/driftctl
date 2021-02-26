# IaC source

Currently, driftctl only supports reading IaC from a Terraform state.
We are investigating to support the Terraform code as well, as a state does not represent an intention.

ℹ️ Multiple states can be read by passing `--from` flags

Example:
```shell
# I want to read a local state and a state stored in an S3 bucket :
driftctl scan \
   --from tfstate+s3://statebucketdriftctl/terraform.tfstate \
   --from tfstate://terraform_toto.tfstate

# You can also use every file under a given prefix for S3
driftctl scan --from tfstate+s3://statebucketdriftctl/states

# ... or in a given local folder
# driftctl will recursively use all files under this folder.
#
# N.B. Symlinks under the root folder will be ignored.
#      If the folder itself is a symlink it will be followed.
driftctl scan --from tfstate://my-states/directory
```


## Supported IaC sources

* Terraform state
  * Local: `--from tfstate://terraform.tfstate`
  * S3: `--from tfstate+s3://my-bucket/path/to/state.tfstate`

### S3

driftctl needs read-only access so you could use the policy below to ensure minimal access to your state file

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
      "Action": "s3:GetObject",
      "Resource": "arn:aws:s3:::mybucket/path/to/my/key"
    }
  ]
}
```
