---
id: authentication
title: Authentication
---

To use driftctl, we need credentials to make authenticated requests to AWS. Just like the AWS CLI, we use [credentials and configuration](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html) settings declared as user environment variables, or in local AWS configuration files.

driftctl supports [named profile](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-profiles.html). By default, the CLI uses the settings found in the profile named `default`. You can override an individual setting by declaring the supported environment variables such as `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_PROFILE` ...

If you are using an [IAM role](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-role.html) as an authorization tool, which is considered a good practice, please be aware that you can still use driftctl by defining a profile for the role in your `~/.aws/config` file.

```shell
[profile driftctlrole]
role_arn = arn:aws:iam::123456789012:role/<NAMEOFTHEROLE>
source_profile = user # profile to assume the role
region = eu-west-3
```

You can now use driftctl by overriding the profile setting.

```shell
$ AWS_PROFILE=driftctlrole driftctl scan
```

## CloudFormation template

Deploy this CloudFormation template to create our limited permission role that you can use as per our above authentication guide.

[![Launch Stack](https://cdn.rawgit.com/buildkite/cloudformation-launch-stack-button-svg/master/launch-stack.svg)](https://console.aws.amazon.com/cloudformation/home?#/stacks/quickcreate?stackName=driftctl-stack&templateURL=https://driftctl-cfn-templates.s3.eu-west-3.amazonaws.com/driftctl-role.yml)

### Update the CloudFormation template

It does not exist an automatic way to update the CloudFormation template from our side because you launched this template on your AWS account. That's why you must be the one to update the template to be on the most recent driftctl role.

Find below two ways to update the CloudFormation template:

1. With the AWS console

- In the [AWS CloudFormation console](https://console.aws.amazon.com/cloudformation), from the list of stacks, select the driftctl stack
- In the stack details pane, choose **Update**
- Select **Replace current template** and specify our **Amazon S3 URL** `https://driftctl-cfn-templates.s3.eu-west-3.amazonaws.com/driftctl-role.yml`, click **Next**
- On the **Specify stack details** and the **Configure stack options** pages, click **Next**
- In the **Change set preview** section, check that AWS CloudFormation will indeed make changes
- Since our template contains one IAM resource, select **I acknowledge that this template may create IAM resources**
- Finally, click **Update stack**

2. With the AWS CLI

```shell
$ aws cloudformation update-stack --stack-name DRIFTCTL_STACK_NAME --template-url https://driftctl-cfn-templates.s3.eu-west-3.amazonaws.com/driftctl-role.yml --capabilities CAPABILITY_NAMED_IAM
```

## Least privileged policy

driftctl needs access to your cloud provider account so that it can list resources on your behalf.

As AWS documentation recommends, the below policy is granting only the permissions required to perform driftctl's tasks.

```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Resource": "*",
            "Action": [
                "ec2:DescribeAddresses",
                "ec2:DescribeImages",
                "ec2:DescribeInstanceAttribute",
                "ec2:DescribeInstances",
                "ec2:DescribeInstanceCreditSpecifications",
                "ec2:DescribeInternetGateways",
                "ec2:DescribeKeyPairs",
                "ec2:DescribeNetworkAcls",
                "ec2:DescribeRouteTables",
                "ec2:DescribeSecurityGroups",
                "ec2:DescribeSnapshots",
                "ec2:DescribeTags",
                "ec2:DescribeVolumes",
                "ec2:DescribeVpcs",
                "ec2:DescribeVpcAttribute",
                "ec2:DescribeVpcClassicLink",
                "ec2:DescribeVpcClassicLinkDnsSupport",
                "ec2:DescribeSubnets",
                "ec2:DescribeNatGateways",
                "iam:GetPolicy",
                "iam:GetPolicyVersion",
                "iam:GetRole",
                "iam:GetRolePolicy",
                "iam:GetUser",
                "iam:GetUserPolicy",
                "iam:ListAccessKeys",
                "iam:ListAttachedRolePolicies",
                "iam:ListAttachedUserPolicies",
                "iam:ListPolicies",
                "iam:ListRolePolicies",
                "iam:ListRoles",
                "iam:ListUserPolicies",
                "iam:ListUsers",
                "lambda:GetFunction",
                "lambda:GetFunctionCodeSigningConfig",
                "lambda:ListFunctions",
                "lambda:ListVersionsByFunction",
                "rds:DescribeDBInstances",
                "rds:DescribeDBSubnetGroups",
                "rds:ListTagsForResource",
                "route53:GetHostedZone",
                "route53:ListHostedZones",
                "route53:ListResourceRecordSets",
                "route53:ListTagsForResource",
                "s3:GetAccelerateConfiguration",
                "s3:GetAnalyticsConfiguration",
                "s3:GetBucketAcl",
                "s3:GetBucketCORS",
                "s3:GetBucketLocation",
                "s3:GetBucketLogging",
                "s3:GetBucketNotification",
                "s3:GetBucketObjectLockConfiguration",
                "s3:GetBucketPolicy",
                "s3:GetBucketRequestPayment",
                "s3:GetBucketTagging",
                "s3:GetBucketVersioning",
                "s3:GetBucketWebsite",
                "s3:GetEncryptionConfiguration",
                "s3:GetInventoryConfiguration",
                "s3:GetLifecycleConfiguration",
                "s3:GetMetricsConfiguration",
                "s3:GetReplicationConfiguration",
                "s3:ListAllMyBuckets",
                "s3:ListBucket",
                "sqs:GetQueueAttributes",
                "sqs:ListQueueTags",
                "sqs:ListQueues",
                "sns:ListTopics",
                "sns:GetTopicAttributes",
                "sns:ListTagsForResource",
                "sns:ListSubscriptions",
                "sns:ListSubscriptionsByTopic",
                "sns:GetSubscriptionAttributes"
            ]
        }
    ]
}
```
