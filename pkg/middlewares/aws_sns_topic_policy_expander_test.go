package middlewares

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"

	awsresource "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/r3labs/diff/v2"

	"github.com/cloudskiff/driftctl/pkg/resource"
)

func TestAwsSNSTopicPolicyExpander_Execute(t *testing.T) {
	tests := []struct {
		name               string
		resourcesFromState *[]resource.Resource
		expected           *[]resource.Resource
		wantErr            bool
	}{
		{
			name: "Inline policy no attached policy",
			resourcesFromState: &[]resource.Resource{
				&awsresource.AwsSnsTopic{
					Arn:    aws.String("arn"),
					Policy: aws.String("{\"policy\": \"coucou\"}"),
					Id:     "ID",
				},
			},
			expected: &[]resource.Resource{
				&awsresource.AwsSnsTopic{
					Arn:    aws.String("arn"),
					Policy: nil,
					Id:     "ID",
				},
				&awsresource.AwsSnsTopicPolicy{
					Arn:    aws.String("arn"),
					Policy: aws.String("{\"policy\":\"coucou\"}"),
					Id:     "ID",
				},
			},
			wantErr: false,
		},
		{
			name: "No inline policy, attached policy",
			resourcesFromState: &[]resource.Resource{
				&awsresource.AwsSnsTopic{
					Arn:    aws.String("arn"),
					Policy: nil,
					Id:     "ID",
				},
				&awsresource.AwsSnsTopicPolicy{
					Arn:    aws.String("arn"),
					Policy: aws.String("{\"policy\": \"coucou\"}"),
					Id:     "ID",
				},
			},
			expected: &[]resource.Resource{
				&awsresource.AwsSnsTopic{
					Arn:    aws.String("arn"),
					Policy: nil,
					Id:     "ID",
				},
				&awsresource.AwsSnsTopicPolicy{
					Arn:    aws.String("arn"),
					Policy: aws.String("{\"policy\": \"coucou\"}"),
					Id:     "ID",
				},
			},
			wantErr: false,
		},
		{
			name: "inline policy and attached policy",
			resourcesFromState: &[]resource.Resource{
				&awsresource.AwsSnsTopic{
					Arn:    aws.String("arn"),
					Policy: aws.String("{\"policy\": \"coucou\"}"),
					Id:     "ID",
				},
				&awsresource.AwsSnsTopicPolicy{
					Arn:    aws.String("arn"),
					Policy: aws.String("{\"policy\": \"coucou\"}"),
					Id:     "ID",
				},
			},
			expected: &[]resource.Resource{
				&awsresource.AwsSnsTopic{
					Arn:    aws.String("arn"),
					Policy: nil,
					Id:     "ID",
				},
				&awsresource.AwsSnsTopicPolicy{
					Arn:    aws.String("arn"),
					Policy: aws.String("{\"policy\": \"coucou\"}"),
					Id:     "ID",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := AwsSNSTopicPolicyExpander{}
			if err := m.Execute(&[]resource.Resource{}, tt.resourcesFromState); (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
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
