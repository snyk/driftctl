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

func TestAwsALBListenerTransformer_Execute(t *testing.T) {
	tests := []struct {
		name               string
		resourcesFromState []*resource.Resource
		mocks              func(*terraform.MockResourceFactory)
		expected           []*resource.Resource
	}{
		{
			name:  "should not transform anything",
			mocks: func(factory *terraform.MockResourceFactory) {},
			resourcesFromState: []*resource.Resource{
				{
					Id:    "foo",
					Type:  aws.AwsS3BucketResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:    "bar",
					Type:  aws.AwsLoadBalancerListenerResourceType,
					Attrs: &resource.Attributes{},
				},
			},
			expected: []*resource.Resource{
				{
					Id:    "foo",
					Type:  aws.AwsS3BucketResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:    "bar",
					Type:  aws.AwsLoadBalancerListenerResourceType,
					Attrs: &resource.Attributes{},
				},
			},
		},
		{
			name: "should transform ALB into LB",
			mocks: func(factory *terraform.MockResourceFactory) {
				factory.
					On("CreateAbstractResource", aws.AwsLoadBalancerListenerResourceType, "alb-test", map[string]interface{}{}).
					Return(&resource.Resource{
						Id:    "alb-test",
						Type:  aws.AwsLoadBalancerListenerResourceType,
						Attrs: &resource.Attributes{},
					}).
					Once()
			},
			resourcesFromState: []*resource.Resource{
				{
					Id:   "foo",
					Type: aws.AwsApiGatewayRestApiResourceType,
					Attrs: &resource.Attributes{
						"body": "{\"info\":{\"title\":\"example\",\"version\":\"1.0\"},\"openapi\":\"3.0.1\",\"paths\":{\"/path1\":{\"get\":{\"parameters\":[{\"in\":\"query\",\"name\":\"type\",\"schema\":{\"type\":\"string\"}},{\"in\":\"query\",\"name\":\"page\",\"schema\":{\"type\":\"string\"}}],\"responses\":{\"200\":{\"content\":{\"application/json\":{\"schema\":{\"$ref\":\"#/components/schemas/Pets\"}}},\"description\":\"200 response\",\"headers\":{\"Access-Control-Allow-Origin\":{\"schema\":{\"type\":\"string\"}}}}},\"x-amazon-apigateway-integration\":{\"httpMethod\":\"GET\",\"payloadFormatVersion\":\"1.0\",\"type\":\"HTTP_PROXY\",\"uri\":\"https://ip-ranges.amazonaws.com/ip-ranges.json\",\"responses\":{\"2\\\\d{2}\":{\"responseTemplates\":{\"application/json\":\"#set ($root=$input.path('$')) { \\\"stage\\\": \\\"$root.name\\\", \\\"user-id\\\": \\\"$root.key\\\" }\",\"application/xml\":\"#set ($root=$input.path('$')) \\u003cstage\\u003e$root.name\\u003c/stage\\u003e \"},\"statusCode\":\"200\"}}}}},\"/path1/path2\":{\"get\":{\"x-amazon-apigateway-integration\":{\"httpMethod\":\"GET\",\"payloadFormatVersion\":\"1.0\",\"type\":\"HTTP_PROXY\",\"uri\":\"https://ip-ranges.amazonaws.com/ip-ranges.json\"}}}}}",
					},
				},
				{
					Id:    "bar",
					Type:  aws.AwsLoadBalancerListenerResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:    "alb-test",
					Type:  aws.AwsApplicationLoadBalancerListenerResourceType,
					Attrs: &resource.Attributes{},
				},
			},
			expected: []*resource.Resource{
				{
					Id:   "foo",
					Type: aws.AwsApiGatewayRestApiResourceType,
					Attrs: &resource.Attributes{
						"body": "{\"info\":{\"title\":\"example\",\"version\":\"1.0\"},\"openapi\":\"3.0.1\",\"paths\":{\"/path1\":{\"get\":{\"parameters\":[{\"in\":\"query\",\"name\":\"type\",\"schema\":{\"type\":\"string\"}},{\"in\":\"query\",\"name\":\"page\",\"schema\":{\"type\":\"string\"}}],\"responses\":{\"200\":{\"content\":{\"application/json\":{\"schema\":{\"$ref\":\"#/components/schemas/Pets\"}}},\"description\":\"200 response\",\"headers\":{\"Access-Control-Allow-Origin\":{\"schema\":{\"type\":\"string\"}}}}},\"x-amazon-apigateway-integration\":{\"httpMethod\":\"GET\",\"payloadFormatVersion\":\"1.0\",\"type\":\"HTTP_PROXY\",\"uri\":\"https://ip-ranges.amazonaws.com/ip-ranges.json\",\"responses\":{\"2\\\\d{2}\":{\"responseTemplates\":{\"application/json\":\"#set ($root=$input.path('$')) { \\\"stage\\\": \\\"$root.name\\\", \\\"user-id\\\": \\\"$root.key\\\" }\",\"application/xml\":\"#set ($root=$input.path('$')) \\u003cstage\\u003e$root.name\\u003c/stage\\u003e \"},\"statusCode\":\"200\"}}}}},\"/path1/path2\":{\"get\":{\"x-amazon-apigateway-integration\":{\"httpMethod\":\"GET\",\"payloadFormatVersion\":\"1.0\",\"type\":\"HTTP_PROXY\",\"uri\":\"https://ip-ranges.amazonaws.com/ip-ranges.json\"}}}}}",
					},
				},
				{
					Id:    "bar",
					Type:  aws.AwsLoadBalancerListenerResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:    "alb-test",
					Type:  aws.AwsLoadBalancerListenerResourceType,
					Attrs: &resource.Attributes{},
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

			m := NewAwsALBListenerTransformer(factory)
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
