package middlewares

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/r3labs/diff/v2"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
)

func TestAwsIamPolicyAttachmentDefaults_Execute(t *testing.T) {
	tests := []struct {
		name               string
		remoteResources    []resource.Resource
		resourcesFromState []resource.Resource
		expected           diff.Changelog
	}{
		{
			"test that default iam policy attachment are excluded when not managed by IaC",
			[]resource.Resource{
				&aws.AwsIamRole{
					Id:   "custom-role",
					Path: func(p string) *string { return &p }("/not-aws-service-role/sso.amazonaws.com"),
				},
				&aws.AwsIamRole{
					Id:   "AWSServiceRoleForSSO",
					Path: func(p string) *string { return &p }("/aws-service-role/sso.amazonaws.com"),
				},
				&aws.AwsIamPolicyAttachment{
					Id:    "driftctl_test-arn:aws:iam::0123456789:policy/driftctl",
					Roles: &[]string{"custom-role"},
				},
				&aws.AwsIamPolicyAttachment{
					Id:    "AWSServiceRoleForSSO-arn:aws:iam::aws:policy/aws-service-role/AWSSSOServiceRolePolicy",
					Roles: &[]string{"AWSServiceRoleForSSO"},
				},
				&aws.AwsIamPolicyAttachment{
					Id:    "AWSServiceRoleForSSO-arn:aws:iam::aws:policy/aws-service-role/whatever",
					Roles: &[]string{"AWSServiceRoleForSSO"},
				},
			},
			[]resource.Resource{},
			diff.Changelog{
				{
					Type: diff.DELETE,
					Path: []string{"0"},
					From: &aws.AwsIamRole{
						Id:   "custom-role",
						Path: func(p string) *string { return &p }("/not-aws-service-role/sso.amazonaws.com"),
					},
					To: nil,
				},
				{
					Type: diff.DELETE,
					Path: []string{"1"},
					From: &aws.AwsIamRole{
						Id:   "AWSServiceRoleForSSO",
						Path: func(p string) *string { return &p }("/aws-service-role/sso.amazonaws.com"),
					},
					To: nil,
				},
				{
					Type: diff.DELETE,
					Path: []string{"2"},
					From: &aws.AwsIamPolicyAttachment{
						Id:    "driftctl_test-arn:aws:iam::0123456789:policy/driftctl",
						Roles: &[]string{"custom-role"},
					},
					To: nil,
				},
			},
		},
		{
			"test that default iam policy attachment are excluded when not managed by IaC",
			[]resource.Resource{
				&aws.AwsIamRole{
					Id:   "custom-role",
					Path: func(p string) *string { return &p }("/not-aws-service-role/sso.amazonaws.com"),
				},
				&aws.AwsIamRole{
					Id:   "AWSServiceRoleForSSO",
					Path: func(p string) *string { return &p }("/aws-service-role/sso.amazonaws.com"),
				},
				&aws.AwsIamPolicyAttachment{
					Id:    "driftctl_test-arn:aws:iam::0123456789:policy/driftctl",
					Roles: &[]string{"custom-role"},
				},
				&aws.AwsIamPolicyAttachment{
					Id:    "AWSServiceRoleForSSO-arn:aws:iam::aws:policy/aws-service-role/AWSSSOServiceRolePolicy",
					Roles: &[]string{"AWSServiceRoleForSSO"},
				},
				&aws.AwsIamPolicyAttachment{
					Id:    "AWSServiceRoleForSSO-arn:aws:iam::aws:policy/aws-service-role/whatever",
					Roles: &[]string{"AWSServiceRoleForSSO", "custom-role"},
				},
			},
			[]resource.Resource{
				&aws.AwsIamPolicyAttachment{
					Id:    "AWSServiceRoleForSSO-arn:aws:iam::aws:policy/aws-service-role/AWSSSOServiceRolePolicy",
					Roles: &[]string{"AWSServiceRoleForSSO"},
				},
			},
			diff.Changelog{
				{
					Type: diff.DELETE,
					Path: []string{"0"},
					From: &aws.AwsIamRole{
						Id:   "custom-role",
						Path: func(p string) *string { return &p }("/not-aws-service-role/sso.amazonaws.com"),
					},
					To: nil,
				},
				{
					Type: diff.DELETE,
					Path: []string{"1"},
					From: &aws.AwsIamRole{
						Id:   "AWSServiceRoleForSSO",
						Path: func(p string) *string { return &p }("/aws-service-role/sso.amazonaws.com"),
					},
					To: nil,
				},
				{
					Type: diff.DELETE,
					Path: []string{"2"},
					From: &aws.AwsIamPolicyAttachment{
						Id:    "driftctl_test-arn:aws:iam::0123456789:policy/driftctl",
						Roles: &[]string{"custom-role"},
					},
					To: nil,
				},
				{
					Type: diff.DELETE,
					Path: []string{"4"},
					From: &aws.AwsIamPolicyAttachment{
						Id:    "AWSServiceRoleForSSO-arn:aws:iam::aws:policy/aws-service-role/whatever",
						Roles: &[]string{"AWSServiceRoleForSSO", "custom-role"},
					},
					To: nil,
				},
			},
		},
		{
			"do not ignore default iam policy attachment when role cannot be found",
			[]resource.Resource{
				&aws.AwsIamRole{
					Id:   "custom-role",
					Path: func(p string) *string { return &p }("/not-aws-service-role/sso.amazonaws.com"),
				},
				&aws.AwsIamPolicyAttachment{
					Id:    "driftctl_test-arn:aws:iam::0123456789:policy/driftctl",
					Roles: &[]string{"custom-role"},
				},
				&aws.AwsIamPolicyAttachment{
					Id:    "AWSServiceRoleForSSO-arn:aws:iam::aws:policy/aws-service-role/AWSSSOServiceRolePolicy",
					Roles: &[]string{"AWSServiceRoleForSSO"},
				},
			},
			[]resource.Resource{},
			diff.Changelog{
				{
					Type: diff.DELETE,
					Path: []string{"0"},
					From: &aws.AwsIamRole{
						Id:   "custom-role",
						Path: func(p string) *string { return &p }("/not-aws-service-role/sso.amazonaws.com"),
					},
					To: nil,
				},
				{
					Type: diff.DELETE,
					Path: []string{"1"},
					From: &aws.AwsIamPolicyAttachment{
						Id:    "driftctl_test-arn:aws:iam::0123456789:policy/driftctl",
						Roles: &[]string{"custom-role"},
					},
					To: nil,
				},
				{
					Type: diff.DELETE,
					Path: []string{"2"},
					From: &aws.AwsIamPolicyAttachment{
						Id:    "AWSServiceRoleForSSO-arn:aws:iam::aws:policy/aws-service-role/AWSSSOServiceRolePolicy",
						Roles: &[]string{"AWSServiceRoleForSSO"},
					},
					To: nil,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := awsIamPolicyAttachmentDefaults(&tt.remoteResources, &tt.resourcesFromState)
			if err != nil {
				t.Fatal(err)
			}

			changelog, err := diff.Diff(tt.remoteResources, tt.resourcesFromState)
			if err != nil {
				t.Fatal(err)
			}

			diffs, err := diff.Diff(tt.expected, changelog)
			if err != nil {
				t.Fatal(err)
			}
			if len(diffs) == 0 {
				return
			}

			for _, change := range diffs {
				t.Errorf("%s got = %v, want %v", strings.Join(change.Path, "."), awsutil.Prettify(change.From), awsutil.Prettify(change.To))
			}
		})
	}
}
