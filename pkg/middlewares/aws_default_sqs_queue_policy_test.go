package middlewares

import (
	"strings"
	"testing"

	awssdk "github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/r3labs/diff/v2"
)

func TestAwsDefaultSqsQueuePolicy_Execute(t *testing.T) {
	tests := []struct {
		name               string
		remoteResources    []resource.Resource
		resourcesFromState []resource.Resource
		expected           []resource.Resource
	}{
		{
			"test default sqs queue policy managed by IaC",
			[]resource.Resource{
				&aws.AwsSqsQueuePolicy{
					Id:     "non-default-sqs-queue-policy",
					Policy: awssdk.String("foo"),
				},
				&aws.AwsSqsQueuePolicy{
					Id:     "default-sqs-queue-policy",
					Policy: awssdk.String(""),
				},
			},
			[]resource.Resource{
				&aws.AwsSqsQueuePolicy{
					Id:     "non-default-sqs-queue-policy",
					Policy: awssdk.String("foo"),
				},
				&aws.AwsSqsQueuePolicy{
					Id:     "default-sqs-queue-policy",
					Policy: awssdk.String(""),
				},
			},
			[]resource.Resource{
				&aws.AwsSqsQueuePolicy{
					Id:     "non-default-sqs-queue-policy",
					Policy: awssdk.String("foo"),
				},
				&aws.AwsSqsQueuePolicy{
					Id:     "default-sqs-queue-policy",
					Policy: awssdk.String(""),
				},
			},
		},
		{
			"test default sqs queue policy not managed by IaC",
			[]resource.Resource{
				&aws.AwsSqsQueuePolicy{
					Id:     "non-default-sqs-queue-policy",
					Policy: awssdk.String("foo"),
				},
				&aws.AwsSqsQueuePolicy{
					Id:     "default-sqs-queue-policy",
					Policy: awssdk.String(""),
				},
			},
			[]resource.Resource{
				&aws.AwsSqsQueuePolicy{
					Id:     "non-default-sqs-queue-policy",
					Policy: awssdk.String("foo"),
				},
			},
			[]resource.Resource{
				&aws.AwsSqsQueuePolicy{
					Id:     "non-default-sqs-queue-policy",
					Policy: awssdk.String("foo"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewAwsDefaultSqsQueuePolicy()
			err := m.Execute(&tt.remoteResources, &tt.resourcesFromState)
			if err != nil {
				t.Fatal(err)
			}

			changelog, err := diff.Diff(tt.remoteResources, tt.expected)
			if err != nil {
				t.Fatal(err)
			}
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s got = %v, want %v", strings.Join(change.Path, "."), awsutil.Prettify(change.From), awsutil.Prettify(change.To))
				}
			}

		})
	}
}
