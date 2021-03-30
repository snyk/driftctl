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
				&aws.AwsIamPolicyAttachment{
					Id: "AWSServiceRoleForSSO-arn:aws:iam::aws:policy/aws-service-role/AWSSSOServiceRolePolicy",
				},
				&aws.AwsIamPolicyAttachment{
					Id: "driftctl_test-arn:aws:iam::0123456789:policy/driftctl",
				},
			},
			[]resource.Resource{
				&aws.AwsIamPolicyAttachment{
					Id: "driftctl_test-arn:aws:iam::0123456789:policy/driftctl",
				},
			},
			diff.Changelog{},
		},
		{
			"test that default iam policy attachment are not excluded when managed by IaC",
			[]resource.Resource{
				&aws.AwsIamPolicyAttachment{
					Id: "AWSServiceRoleForSSO-arn:aws:iam::aws:policy/aws-service-role/AWSSSOServiceRolePolicy",
				},
				&aws.AwsIamPolicyAttachment{
					Id: "driftctl_test-arn:aws:iam::0123456789:policy/driftctl",
				},
			},
			[]resource.Resource{
				&aws.AwsIamPolicyAttachment{
					Id: "AWSServiceRoleForSSO-arn:aws:iam::aws:policy/aws-service-role/AWSSSOServiceRolePolicy",
				},
			},
			diff.Changelog{
				{
					Type: "delete",
					Path: []string{"1"},
					From: &aws.AwsIamPolicyAttachment{
						Id: "driftctl_test-arn:aws:iam::0123456789:policy/driftctl",
					},
					To: nil,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewAwsIamPolicyAttachmentDefaults()
			err := m.Execute(&tt.remoteResources, &tt.resourcesFromState)
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
