package middlewares

import (
	"strings"
	"testing"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/cloudskiff/driftctl/pkg/terraform"

	"github.com/r3labs/diff/v2"
)

func TestAwsBucketPolicyExpander_Execute(t *testing.T) {
	tests := []struct {
		name               string
		resourcesFromState []*resource.Resource
		mocks              func(*terraform.MockResourceFactory)
		expected           []*resource.Resource
	}{
		{
			name: "Inline policy, no aws_s3_bucket_policy attached",
			mocks: func(factory *terraform.MockResourceFactory) {
				factory.On(
					"CreateAbstractResource",
					aws.AwsS3BucketPolicyResourceType,
					"foo",
					map[string]interface{}{
						"id":     "foo",
						"bucket": "foo",
						"policy": "{\"Id\":\"MYINLINEBUCKETPOLICY\",\"Statement\":[{\"Action\":\"s3:*\",\"Condition\":{\"IpAddress\":{\"aws:SourceIp\":\"8.8.8.8/32\"}},\"Effect\":\"Deny\",\"Principal\":\"*\",\"Resource\":\"arn:aws:s3:::bucket-test-policy-like-sqs/*\",\"Sid\":\"IPAllow\"}],\"Version\":\"2012-10-17\"}",
					},
				).Once().Return(&resource.Resource{
					Id:   "foo",
					Type: aws.AwsS3BucketPolicyResourceType,
				})
			},
			resourcesFromState: []*resource.Resource{
				{
					Id:   "foo",
					Type: aws.AwsS3BucketResourceType,
					Attrs: &resource.Attributes{
						"bucket": "foo",
						"policy": "{\"Id\":\"MYINLINEBUCKETPOLICY\",\"Statement\":[{\"Action\":\"s3:*\",\"Condition\":{\"IpAddress\":{\"aws:SourceIp\":\"8.8.8.8/32\"}},\"Effect\":\"Deny\",\"Principal\":\"*\",\"Resource\":\"arn:aws:s3:::bucket-test-policy-like-sqs/*\",\"Sid\":\"IPAllow\"}],\"Version\":\"2012-10-17\"}",
					},
				},
			},
			expected: []*resource.Resource{
				{
					Id:   "foo",
					Type: aws.AwsS3BucketResourceType,
					Attrs: &resource.Attributes{
						"bucket": "foo",
					},
				},
				{
					Id:   "foo",
					Type: aws.AwsS3BucketPolicyResourceType,
				},
			},
		},
		{
			name: "No inline policy, aws_s3_bucket_policy attached",
			mocks: func(factory *terraform.MockResourceFactory) {
				factory.On(
					"CreateAbstractResource",
					aws.AwsS3BucketPolicyResourceType,
					"foo",
					map[string]interface{}{
						"id":     "foo",
						"bucket": "foo",
						"policy": "{\"Id\":\"MYBUCKETPOLICY\",\"Statement\":[{\"Action\":\"s3:*\",\"Condition\":{\"IpAddress\":{\"aws:SourceIp\":\"8.8.8.8/32\"}},\"Effect\":\"Deny\",\"Principal\":\"*\",\"Resource\":\"arn:aws:s3:::bucket-test-policy-like-sqs/*\",\"Sid\":\"IPAllow\"}],\"Version\":\"2012-10-17\"}",
					},
				).Once().Return(&resource.Resource{
					Id:   "foo",
					Type: aws.AwsS3BucketPolicyResourceType,
				})
			},
			resourcesFromState: []*resource.Resource{
				{
					Id:   "foo",
					Type: aws.AwsS3BucketResourceType,
					Attrs: &resource.Attributes{
						"bucket": "foo",
					},
				},
				{
					Id:   "foo",
					Type: aws.AwsS3BucketPolicyResourceType,
				},
			},
			expected: []*resource.Resource{
				{
					Id:   "foo",
					Type: aws.AwsS3BucketResourceType,
					Attrs: &resource.Attributes{
						"bucket": "foo",
					},
				},
				{
					Id:   "foo",
					Type: aws.AwsS3BucketPolicyResourceType,
				},
			},
		},
		{
			name: "Inline policy and aws_s3_bucket_policy",
			resourcesFromState: []*resource.Resource{
				{
					Id:   "foo",
					Type: aws.AwsS3BucketResourceType,
					Attrs: &resource.Attributes{
						"bucket": "foo",
						"policy": awssdk.String("{\"Id\":\"MYINLINEBUCKETPOLICY\",\"Statement\":[{\"Action\":\"s3:*\",\"Condition\":{\"IpAddress\":{\"aws:SourceIp\":\"8.8.8.8/32\"}},\"Effect\":\"Deny\",\"Principal\":\"*\",\"Resource\":\"arn:aws:s3:::bucket-test-policy-like-sqs/*\",\"Sid\":\"IPAllow\"}],\"Version\":\"2012-10-17\"}"),
					},
				},
				{
					Id:   "foo",
					Type: aws.AwsS3BucketPolicyResourceType,
					Attrs: &resource.Attributes{
						"bucket": "foo",
						"policy": "{\"Id\":\"MYBUCKETPOLICY\",\"Statement\":[{\"Action\":\"s3:*\",\"Condition\":{\"IpAddress\":{\"aws:SourceIp\":\"8.8.8.8/32\"}},\"Effect\":\"Deny\",\"Principal\":\"*\",\"Resource\":\"arn:aws:s3:::bucket-test-policy-like-sqs/*\",\"Sid\":\"IPAllow\"}],\"Version\":\"2012-10-17\"}",
					},
				},
			},
			expected: []*resource.Resource{
				{
					Id:   "foo",
					Type: aws.AwsS3BucketResourceType,
					Attrs: &resource.Attributes{
						"bucket": "foo",
					},
				},
				{
					Id:   "foo",
					Type: aws.AwsS3BucketPolicyResourceType,
					Attrs: &resource.Attributes{
						"bucket": "foo",
						"policy": "{\"Id\":\"MYBUCKETPOLICY\",\"Statement\":[{\"Action\":\"s3:*\",\"Condition\":{\"IpAddress\":{\"aws:SourceIp\":\"8.8.8.8/32\"}},\"Effect\":\"Deny\",\"Principal\":\"*\",\"Resource\":\"arn:aws:s3:::bucket-test-policy-like-sqs/*\",\"Sid\":\"IPAllow\"}],\"Version\":\"2012-10-17\"}",
					},
				},
			},
		},
		{
			name: "empty policy ",
			resourcesFromState: []*resource.Resource{
				{
					Id:   "foo",
					Type: aws.AwsS3BucketResourceType,
					Attrs: &resource.Attributes{
						"bucket": "foo",
						"policy": "",
					},
				},
			},
			expected: []*resource.Resource{
				{
					Id:   "foo",
					Type: aws.AwsS3BucketResourceType,
					Attrs: &resource.Attributes{
						"bucket": "foo",
						"policy": "",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			factory := &terraform.MockResourceFactory{}
			if tt.mocks != nil {
				tt.mocks(factory)
			}

			m := NewAwsBucketPolicyExpander(factory)
			err := m.Execute(&[]*resource.Resource{}, &tt.resourcesFromState)
			if err != nil {
				t.Fatal(err)
			}
			changelog, err := diff.Diff(tt.expected, tt.resourcesFromState)
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
