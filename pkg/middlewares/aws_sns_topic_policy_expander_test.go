package middlewares

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/stretchr/testify/mock"
	"github.com/zclconf/go-cty/cty"

	awsresource "github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	testresource "github.com/cloudskiff/driftctl/test/resource"

	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/r3labs/diff/v2"

	"github.com/cloudskiff/driftctl/pkg/resource"
)

func TestAwsSNSTopicPolicyExpander_Execute(t *testing.T) {
	tests := []struct {
		name               string
		resourcesFromState *[]resource.Resource
		expected           *[]resource.Resource
		mock               func(factory *terraform.MockResourceFactory)
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
				&resource.AbstractResource{
					Id:   "ID",
					Type: awsresource.AwsSnsTopicPolicyResourceType,
					Attrs: &resource.Attributes{
						"arn":    "arn",
						"id":     "ID",
						"policy": "{\"policy\":\"coucou\"}",
					},
				},
			},
			mock: func(factory *terraform.MockResourceFactory) {
				foo := cty.ObjectVal(map[string]cty.Value{
					"arn":    cty.StringVal("arn"),
					"id":     cty.StringVal("ID"),
					"policy": cty.StringVal("{\"policy\":\"coucou\"}"),
				})

				factory.On("CreateResource", mock.MatchedBy(func(input map[string]interface{}) bool {
					return input["id"] == "ID"
				}), awsresource.AwsSnsTopicPolicyResourceType).Once().Return(&foo, nil)
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
				&resource.AbstractResource{
					Id:   "ID",
					Type: awsresource.AwsSnsTopicPolicyResourceType,
					Attrs: &resource.Attributes{
						"arn":    "arn",
						"id":     "ID",
						"policy": "{\"policy\":\"coucou\"}",
					},
				},
			},
			expected: &[]resource.Resource{
				&awsresource.AwsSnsTopic{
					Arn:    aws.String("arn"),
					Policy: nil,
					Id:     "ID",
				},
				&resource.AbstractResource{
					Id:   "ID",
					Type: awsresource.AwsSnsTopicPolicyResourceType,
					Attrs: &resource.Attributes{
						"arn":    "arn",
						"id":     "ID",
						"policy": "{\"policy\":\"coucou\"}",
					},
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

			factory := &terraform.MockResourceFactory{}
			if tt.mock != nil {
				tt.mock(factory)
			}

			repo := testresource.InitFakeSchemaRepository("aws", "3.19.0")
			awsresource.InitResourcesMetadata(repo)

			m := NewAwsSNSTopicPolicyExpander(factory, repo)
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
