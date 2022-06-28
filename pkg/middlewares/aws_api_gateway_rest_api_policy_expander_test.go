package middlewares

import (
	"github.com/snyk/driftctl/enumeration/terraform"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/r3labs/diff/v2"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/aws"
)

func TestAwsApiGatewayRestApiPolicyPolicyExpander_Execute(t *testing.T) {
	tests := []struct {
		name               string
		resourcesFromState []*resource.Resource
		mocks              func(*terraform.MockResourceFactory)
		expected           []*resource.Resource
	}{
		{
			name: "Inline policy, no aws_api_gateway_rest_api_policy attached",
			mocks: func(factory *terraform.MockResourceFactory) {
				factory.On(
					"CreateAbstractResource",
					aws.AwsApiGatewayRestApiPolicyResourceType,
					"foo",
					map[string]interface{}{
						"id":          "foo",
						"rest_api_id": "foo",
						"policy":      "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":\"*\",\"Action\":\"execute-api:Invoke\",\"Resource\":\"arn:aws:execute-api:us-east-1:011111111111:rrwhncu4h2/*\"}]}",
					},
				).Once().Return(&resource.Resource{
					Id:   "foo",
					Type: aws.AwsApiGatewayRestApiPolicyResourceType,
				})
			},
			resourcesFromState: []*resource.Resource{
				{
					Id:   "foo",
					Type: aws.AwsApiGatewayRestApiResourceType,
					Attrs: &resource.Attributes{
						"policy": "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":\"*\",\"Action\":\"execute-api:Invoke\",\"Resource\":\"arn:aws:execute-api:us-east-1:011111111111:rrwhncu4h2/*\"}]}",
					},
				},
			},
			expected: []*resource.Resource{
				{
					Id:    "foo",
					Type:  aws.AwsApiGatewayRestApiResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:   "foo",
					Type: aws.AwsApiGatewayRestApiPolicyResourceType,
				},
			},
		},
		{
			name: "No inline policy, aws_api_gateway_rest_api_policy attached",
			resourcesFromState: []*resource.Resource{
				{
					Id:    "foo",
					Type:  aws.AwsApiGatewayRestApiResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:   "foo",
					Type: aws.AwsApiGatewayRestApiPolicyResourceType,
				},
			},
			expected: []*resource.Resource{
				{
					Id:    "foo",
					Type:  aws.AwsApiGatewayRestApiResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:   "foo",
					Type: aws.AwsApiGatewayRestApiPolicyResourceType,
				},
			},
		},
		{
			name: "Inline policy and aws_api_gateway_rest_api_policy",
			resourcesFromState: []*resource.Resource{
				{
					Id:   "foo",
					Type: aws.AwsApiGatewayRestApiResourceType,
					Attrs: &resource.Attributes{
						"policy": "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":\"*\",\"Action\":\"execute-api:Invoke\",\"Resource\":\"arn:aws:execute-api:us-east-1:011111111111:rrwhncu4h2/*\"}]}",
					},
				},
				{
					Id:   "foo",
					Type: aws.AwsApiGatewayRestApiPolicyResourceType,
				},
			},
			expected: []*resource.Resource{
				{
					Id:    "foo",
					Type:  aws.AwsApiGatewayRestApiResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:   "foo",
					Type: aws.AwsApiGatewayRestApiPolicyResourceType,
				},
			},
		},
		{
			name: "empty policy",
			resourcesFromState: []*resource.Resource{
				{
					Id:   "foo",
					Type: aws.AwsApiGatewayRestApiResourceType,
					Attrs: &resource.Attributes{
						"policy": "",
					},
				},
			},
			expected: []*resource.Resource{
				{
					Id:   "foo",
					Type: aws.AwsApiGatewayRestApiResourceType,
					Attrs: &resource.Attributes{
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

			m := NewAwsApiGatewayRestApiPolicyExpander(factory)
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
