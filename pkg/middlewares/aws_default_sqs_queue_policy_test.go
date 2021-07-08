package middlewares

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/r3labs/diff/v2"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
)

func TestAwsDefaultSQSQueuePolicy_Execute(t *testing.T) {
	tests := []struct {
		name               string
		remoteResources    []resource.Resource
		resourcesFromState []resource.Resource
		expected           []resource.Resource
	}{
		{
			"test default sqs queue policy managed by IaC",
			[]resource.Resource{
				&resource.AbstractResource{
					Id:   "non-default-sqs-queue-policy",
					Type: aws.AwsSqsQueuePolicyResourceType,
					Attrs: &resource.Attributes{
						"queue_url": "non-default-sqs-queue-policy",
						"id":        "non-default-sqs-queue-policy",
						"policy":    "foo",
					},
				},
				&resource.AbstractResource{
					Id:   "default-sqs-queue-policy",
					Type: aws.AwsSqsQueuePolicyResourceType,
					Attrs: &resource.Attributes{
						"queue_url": "default-sqs-queue-policy",
						"id":        "default-sqs-queue-policy",
						"policy":    "",
					},
				},
			},
			[]resource.Resource{
				&resource.AbstractResource{
					Id:   "non-default-sqs-queue-policy",
					Type: aws.AwsSqsQueuePolicyResourceType,
					Attrs: &resource.Attributes{
						"queue_url": "non-default-sqs-queue-policy",
						"id":        "non-default-sqs-queue-policy",
						"policy":    "foo",
					},
				},
				&resource.AbstractResource{
					Id:   "default-sqs-queue-policy",
					Type: aws.AwsSqsQueuePolicyResourceType,
					Attrs: &resource.Attributes{
						"queue_url": "default-sqs-queue-policy",
						"id":        "default-sqs-queue-policy",
						"policy":    "",
					},
				},
			},
			[]resource.Resource{
				&resource.AbstractResource{
					Id:   "non-default-sqs-queue-policy",
					Type: aws.AwsSqsQueuePolicyResourceType,
					Attrs: &resource.Attributes{
						"queue_url": "non-default-sqs-queue-policy",
						"id":        "non-default-sqs-queue-policy",
						"policy":    "foo",
					},
				},
				&resource.AbstractResource{
					Id:   "default-sqs-queue-policy",
					Type: aws.AwsSqsQueuePolicyResourceType,
					Attrs: &resource.Attributes{
						"queue_url": "default-sqs-queue-policy",
						"id":        "default-sqs-queue-policy",
						"policy":    "",
					},
				},
			},
		},
		{
			"test default sqs queue policy not managed by IaC",
			[]resource.Resource{
				&resource.AbstractResource{
					Id:   "non-default-sqs-queue-policy",
					Type: aws.AwsSqsQueuePolicyResourceType,
					Attrs: &resource.Attributes{
						"queue_url": "non-default-sqs-queue-policy",
						"id":        "non-default-sqs-queue-policy",
						"policy":    "foo",
					},
				},
				&resource.AbstractResource{
					Id:   "default-sqs-queue-policy",
					Type: aws.AwsSqsQueuePolicyResourceType,
					Attrs: &resource.Attributes{
						"queue_url": "default-sqs-queue-policy",
						"id":        "default-sqs-queue-policy",
						"policy":    "",
					},
				},
			},
			[]resource.Resource{
				&resource.AbstractResource{
					Id:   "non-default-sqs-queue-policy",
					Type: aws.AwsSqsQueuePolicyResourceType,
					Attrs: &resource.Attributes{
						"queue_url": "non-default-sqs-queue-policy",
						"id":        "non-default-sqs-queue-policy",
						"policy":    "foo",
					},
				},
			},
			[]resource.Resource{
				&resource.AbstractResource{
					Id:   "non-default-sqs-queue-policy",
					Type: aws.AwsSqsQueuePolicyResourceType,
					Attrs: &resource.Attributes{
						"queue_url": "non-default-sqs-queue-policy",
						"id":        "non-default-sqs-queue-policy",
						"policy":    "foo",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewAwsDefaultSQSQueuePolicy()
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
