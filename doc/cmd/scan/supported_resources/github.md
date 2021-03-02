# Github

## Authentication

To use driftctl, we need credentials to make authenticated requests to github. Just like the terraform provider, we retrieve config from [environment variables](https://registry.terraform.io/providers/integrations/github/latest/docs#argument-reference).

```bash
$ GITHUB_TOKEN=14758f1afd44c09b7992073ccf00b43d GITHUB_ORGANIZATION=my-org driftctl scan --to github+tf
```

## Least privileged policy

Below you can find the minimal scope required for driftctl to be able to scan every github supported resources.

```shell
repo # Required to enumerate public and private repos
read:org # Used to list your organization teams
```

**⚠️ Beware that if you don't set correct permissions for your token, you won't see any errors and all resources will appear as deleted from remote**

## Supported resources

- [x] github_repository
- [x] github_team
- [x] github_membership
