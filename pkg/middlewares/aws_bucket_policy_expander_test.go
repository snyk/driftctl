package middlewares

import (
	"strings"
	"testing"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/stretchr/testify/mock"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/cloudskiff/driftctl/pkg/terraform"

	"github.com/r3labs/diff/v2"
)

func TestAwsBucketPolicyExpander_Execute(t *testing.T) {
	tests := []struct {
		name               string
		resourcesFromState []resource.Resource
		expected           []resource.Resource
	}{
		{
			"Inline policy, no aws_s3_bucket_policy attached",
			[]resource.Resource{
				&aws.AwsS3Bucket{
					Id:     "foo",
					Bucket: awssdk.String("foo"),
					Policy: awssdk.String("{\"Id\":\"MYINLINEBUCKETPOLICY\",\"Statement\":[{\"Action\":\"s3:*\",\"Condition\":{\"IpAddress\":{\"aws:SourceIp\":\"8.8.8.8/32\"}},\"Effect\":\"Deny\",\"Principal\":\"*\",\"Resource\":\"arn:aws:s3:::bucket-test-policy-like-sqs/*\",\"Sid\":\"IPAllow\"}],\"Version\":\"2012-10-17\"}"),
				},
			},
			[]resource.Resource{
				&aws.AwsS3Bucket{
					Id:     "foo",
					Bucket: awssdk.String("foo"),
					Policy: nil,
				},
				&aws.AwsS3BucketPolicy{
					Id:     "foo",
					Bucket: awssdk.String("foo"),
					Policy: awssdk.String("{\"Id\":\"MYINLINEBUCKETPOLICY\",\"Statement\":[{\"Action\":\"s3:*\",\"Condition\":{\"IpAddress\":{\"aws:SourceIp\":\"8.8.8.8/32\"}},\"Effect\":\"Deny\",\"Principal\":\"*\",\"Resource\":\"arn:aws:s3:::bucket-test-policy-like-sqs/*\",\"Sid\":\"IPAllow\"}],\"Version\":\"2012-10-17\"}"),
				},
			},
		},
		{
			"No inline policy, aws_s3_bucket_policy attached",
			[]resource.Resource{
				&aws.AwsS3Bucket{
					Id:     "foo",
					Bucket: awssdk.String("foo"),
					Policy: nil,
				},
				&aws.AwsS3BucketPolicy{
					Id:     "foo",
					Bucket: awssdk.String("foo"),
					Policy: awssdk.String("{\"Id\":\"MYBUCKETPOLICY\",\"Statement\":[{\"Action\":\"s3:*\",\"Condition\":{\"IpAddress\":{\"aws:SourceIp\":\"8.8.8.8/32\"}},\"Effect\":\"Deny\",\"Principal\":\"*\",\"Resource\":\"arn:aws:s3:::bucket-test-policy-like-sqs/*\",\"Sid\":\"IPAllow\"}],\"Version\":\"2012-10-17\"}"),
				},
			},
			[]resource.Resource{
				&aws.AwsS3Bucket{
					Id:     "foo",
					Bucket: awssdk.String("foo"),
					Policy: nil,
				},
				&aws.AwsS3BucketPolicy{
					Id:     "foo",
					Bucket: awssdk.String("foo"),
					Policy: awssdk.String("{\"Id\":\"MYBUCKETPOLICY\",\"Statement\":[{\"Action\":\"s3:*\",\"Condition\":{\"IpAddress\":{\"aws:SourceIp\":\"8.8.8.8/32\"}},\"Effect\":\"Deny\",\"Principal\":\"*\",\"Resource\":\"arn:aws:s3:::bucket-test-policy-like-sqs/*\",\"Sid\":\"IPAllow\"}],\"Version\":\"2012-10-17\"}"),
				},
			},
		},
		{
			"Inline policy and aws_s3_bucket_policy",
			[]resource.Resource{
				&aws.AwsS3Bucket{
					Id:     "foo",
					Bucket: awssdk.String("foo"),
					Policy: awssdk.String("{\"Id\":\"MYINLINEBUCKETPOLICY\",\"Statement\":[{\"Action\":\"s3:*\",\"Condition\":{\"IpAddress\":{\"aws:SourceIp\":\"8.8.8.8/32\"}},\"Effect\":\"Deny\",\"Principal\":\"*\",\"Resource\":\"arn:aws:s3:::bucket-test-policy-like-sqs/*\",\"Sid\":\"IPAllow\"}],\"Version\":\"2012-10-17\"}"),
				},
				&aws.AwsS3BucketPolicy{
					Id:     "foo",
					Bucket: awssdk.String("foo"),
					Policy: awssdk.String("{\"Id\":\"MYBUCKETPOLICY\",\"Statement\":[{\"Action\":\"s3:*\",\"Condition\":{\"IpAddress\":{\"aws:SourceIp\":\"8.8.8.8/32\"}},\"Effect\":\"Deny\",\"Principal\":\"*\",\"Resource\":\"arn:aws:s3:::bucket-test-policy-like-sqs/*\",\"Sid\":\"IPAllow\"}],\"Version\":\"2012-10-17\"}"),
				},
			},
			[]resource.Resource{
				&aws.AwsS3Bucket{
					Id:     "foo",
					Bucket: awssdk.String("foo"),
					Policy: nil,
				},
				&aws.AwsS3BucketPolicy{
					Id:     "foo",
					Bucket: awssdk.String("foo"),
					Policy: awssdk.String("{\"Id\":\"MYBUCKETPOLICY\",\"Statement\":[{\"Action\":\"s3:*\",\"Condition\":{\"IpAddress\":{\"aws:SourceIp\":\"8.8.8.8/32\"}},\"Effect\":\"Deny\",\"Principal\":\"*\",\"Resource\":\"arn:aws:s3:::bucket-test-policy-like-sqs/*\",\"Sid\":\"IPAllow\"}],\"Version\":\"2012-10-17\"}"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			factory := &terraform.MockResourceFactory{}
			factory.On("CreateResource", mock.Anything, "aws_s3_bucket_policy").Once().Return(nil, nil)

			m := NewAwsBucketPolicyExpander(factory)
			err := m.Execute(&[]resource.Resource{}, &tt.resourcesFromState)
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
