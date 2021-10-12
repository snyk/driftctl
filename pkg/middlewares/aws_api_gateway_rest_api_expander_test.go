package middlewares

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/cloudskiff/driftctl/pkg/terraform"

	"github.com/r3labs/diff/v2"
)

func TestAwsApiGatewayRestApiExpander_Execute(t *testing.T) {
	tests := []struct {
		name               string
		resourcesFromState []*resource.Resource
		remoteResources    []*resource.Resource
		mocks              func(*terraform.MockResourceFactory)
		expected           []*resource.Resource
	}{
		{
			name: "create aws_api_gateway_resource from OpenAPI v3 document",
			mocks: func(factory *terraform.MockResourceFactory) {
				factory.On(
					"CreateAbstractResource",
					aws.AwsApiGatewayResourceResourceType,
					"bar",
					map[string]interface{}{
						"rest_api_id": "foo",
						"path":        "/path1",
					},
				).Once().Return(&resource.Resource{
					Id:   "bar",
					Type: aws.AwsApiGatewayResourceResourceType,
					Attrs: &resource.Attributes{
						"rest_api_id": "foo",
						"path":        "/path1",
					},
				})
				factory.On(
					"CreateAbstractResource",
					aws.AwsApiGatewayResourceResourceType,
					"baz",
					map[string]interface{}{
						"rest_api_id": "foo",
						"path":        "/path1/path2",
					},
				).Once().Return(&resource.Resource{
					Id:   "baz",
					Type: aws.AwsApiGatewayResourceResourceType,
					Attrs: &resource.Attributes{
						"rest_api_id": "foo",
						"path":        "/path1/path2",
					},
				})
			},
			resourcesFromState: []*resource.Resource{
				{
					Id:   "foo",
					Type: aws.AwsApiGatewayRestApiResourceType,
					Attrs: &resource.Attributes{
						"body": "{\"info\":{\"title\":\"example\",\"version\":\"1.0\"},\"openapi\":\"3.0.1\",\"paths\":{\"/path1\":{\"get\":{\"x-amazon-apigateway-integration\":{\"httpMethod\":\"GET\",\"payloadFormatVersion\":\"1.0\",\"type\":\"HTTP_PROXY\",\"uri\":\"https://ip-ranges.amazonaws.com/ip-ranges.json\"}}},\"/path1/path2\":{\"get\":{\"x-amazon-apigateway-integration\":{\"httpMethod\":\"GET\",\"payloadFormatVersion\":\"1.0\",\"type\":\"HTTP_PROXY\",\"uri\":\"https://ip-ranges.amazonaws.com/ip-ranges.json\"}}}}}",
					},
				},
			},
			remoteResources: []*resource.Resource{
				{
					Id:   "bar",
					Type: aws.AwsApiGatewayResourceResourceType,
					Attrs: &resource.Attributes{
						"rest_api_id": "foo",
						"path":        "/path1",
					},
				},
				{
					Id:   "baz",
					Type: aws.AwsApiGatewayResourceResourceType,
					Attrs: &resource.Attributes{
						"rest_api_id": "foo",
						"path":        "/path1/path2",
					},
				},
			},
			expected: []*resource.Resource{
				{
					Id:   "foo",
					Type: aws.AwsApiGatewayRestApiResourceType,
					Attrs: &resource.Attributes{
						"body": "{\"info\":{\"title\":\"example\",\"version\":\"1.0\"},\"openapi\":\"3.0.1\",\"paths\":{\"/path1\":{\"get\":{\"x-amazon-apigateway-integration\":{\"httpMethod\":\"GET\",\"payloadFormatVersion\":\"1.0\",\"type\":\"HTTP_PROXY\",\"uri\":\"https://ip-ranges.amazonaws.com/ip-ranges.json\"}}},\"/path1/path2\":{\"get\":{\"x-amazon-apigateway-integration\":{\"httpMethod\":\"GET\",\"payloadFormatVersion\":\"1.0\",\"type\":\"HTTP_PROXY\",\"uri\":\"https://ip-ranges.amazonaws.com/ip-ranges.json\"}}}}}",
					},
				},
				{
					Id:   "bar",
					Type: aws.AwsApiGatewayResourceResourceType,
					Attrs: &resource.Attributes{
						"rest_api_id": "foo",
						"path":        "/path1",
					},
				},
				{
					Id:   "baz",
					Type: aws.AwsApiGatewayResourceResourceType,
					Attrs: &resource.Attributes{
						"rest_api_id": "foo",
						"path":        "/path1/path2",
					},
				},
			},
		},
		{
			name: "create aws_api_gateway_resource from OpenAPI v2 document",
			mocks: func(factory *terraform.MockResourceFactory) {
				factory.On(
					"CreateAbstractResource",
					aws.AwsApiGatewayResourceResourceType,
					"bar",
					map[string]interface{}{
						"rest_api_id": "foo",
						"path":        "/test",
					},
				).Once().Return(&resource.Resource{
					Id:   "bar",
					Type: aws.AwsApiGatewayResourceResourceType,
					Attrs: &resource.Attributes{
						"rest_api_id": "foo",
						"path":        "/test",
					},
				})
			},
			resourcesFromState: []*resource.Resource{
				{
					Id:   "foo",
					Type: aws.AwsApiGatewayRestApiResourceType,
					Attrs: &resource.Attributes{
						"body": "{\"info\":{\"title\":\"test\",\"version\":\"2017-04-20T04:08:08Z\"},\"paths\":{\"/test\":{\"get\":{\"responses\":{\"200\":{\"description\":\"OK\"}},\"x-amazon-apigateway-integration\":{\"httpMethod\":\"GET\",\"responses\":{\"default\":{\"statusCode\":200}},\"type\":\"HTTP\",\"uri\":\"https://aws.amazon.com/\"}}}},\"schemes\":[\"https\"],\"swagger\":\"2.0\"}",
					},
				},
			},
			remoteResources: []*resource.Resource{
				{
					Id:   "bar",
					Type: aws.AwsApiGatewayResourceResourceType,
					Attrs: &resource.Attributes{
						"rest_api_id": "foo",
						"path":        "/test",
					},
				},
			},
			expected: []*resource.Resource{
				{
					Id:   "foo",
					Type: aws.AwsApiGatewayRestApiResourceType,
					Attrs: &resource.Attributes{
						"body": "{\"info\":{\"title\":\"test\",\"version\":\"2017-04-20T04:08:08Z\"},\"paths\":{\"/test\":{\"get\":{\"responses\":{\"200\":{\"description\":\"OK\"}},\"x-amazon-apigateway-integration\":{\"httpMethod\":\"GET\",\"responses\":{\"default\":{\"statusCode\":200}},\"type\":\"HTTP\",\"uri\":\"https://aws.amazon.com/\"}}}},\"schemes\":[\"https\"],\"swagger\":\"2.0\"}",
					},
				},
				{
					Id:   "bar",
					Type: aws.AwsApiGatewayResourceResourceType,
					Attrs: &resource.Attributes{
						"rest_api_id": "foo",
						"path":        "/test",
					},
				},
			},
		},
		{
			name: "empty or unknown body",
			resourcesFromState: []*resource.Resource{
				{
					Id:   "foo",
					Type: aws.AwsApiGatewayRestApiResourceType,
					Attrs: &resource.Attributes{
						"body": "",
					},
				},
				{
					Id:    "bar",
					Type:  aws.AwsApiGatewayRestApiResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:   "baz",
					Type: aws.AwsApiGatewayRestApiResourceType,
					Attrs: &resource.Attributes{
						"body": "{}",
					},
				},
			},
			expected: []*resource.Resource{
				{
					Id:   "foo",
					Type: aws.AwsApiGatewayRestApiResourceType,
					Attrs: &resource.Attributes{
						"body": "",
					},
				},
				{
					Id:    "bar",
					Type:  aws.AwsApiGatewayRestApiResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:   "baz",
					Type: aws.AwsApiGatewayRestApiResourceType,
					Attrs: &resource.Attributes{
						"body": "{}",
					},
				},
			},
		},
		{
			name: "create resources with same path but not the same rest api id",
			mocks: func(factory *terraform.MockResourceFactory) {
				factory.On(
					"CreateAbstractResource",
					aws.AwsApiGatewayResourceResourceType,
					"foo-path1",
					map[string]interface{}{
						"rest_api_id": "foo",
						"path":        "/path1",
					},
				).Once().Return(&resource.Resource{
					Id:   "foo-path1",
					Type: aws.AwsApiGatewayResourceResourceType,
					Attrs: &resource.Attributes{
						"rest_api_id": "foo",
						"path":        "/path1",
					},
				})
				factory.On(
					"CreateAbstractResource",
					aws.AwsApiGatewayResourceResourceType,
					"foo-path1-path2",
					map[string]interface{}{
						"rest_api_id": "foo",
						"path":        "/path1/path2",
					},
				).Once().Return(&resource.Resource{
					Id:   "foo-path1-path2",
					Type: aws.AwsApiGatewayResourceResourceType,
					Attrs: &resource.Attributes{
						"rest_api_id": "foo",
						"path":        "/path1/path2",
					},
				})
				factory.On(
					"CreateAbstractResource",
					aws.AwsApiGatewayResourceResourceType,
					"bar-path1",
					map[string]interface{}{
						"rest_api_id": "bar",
						"path":        "/path1",
					},
				).Once().Return(&resource.Resource{
					Id:   "bar-path1",
					Type: aws.AwsApiGatewayResourceResourceType,
					Attrs: &resource.Attributes{
						"rest_api_id": "bar",
						"path":        "/path1",
					},
				})
				factory.On(
					"CreateAbstractResource",
					aws.AwsApiGatewayResourceResourceType,
					"bar-path1-path2",
					map[string]interface{}{
						"rest_api_id": "bar",
						"path":        "/path1/path2",
					},
				).Once().Return(&resource.Resource{
					Id:   "bar-path1-path2",
					Type: aws.AwsApiGatewayResourceResourceType,
					Attrs: &resource.Attributes{
						"rest_api_id": "bar",
						"path":        "/path1/path2",
					},
				})
			},
			resourcesFromState: []*resource.Resource{
				{
					Id:   "foo",
					Type: aws.AwsApiGatewayRestApiResourceType,
					Attrs: &resource.Attributes{
						"body": "{\"info\":{\"title\":\"example\",\"version\":\"1.0\"},\"openapi\":\"3.0.1\",\"paths\":{\"/path1\":{\"get\":{\"x-amazon-apigateway-integration\":{\"httpMethod\":\"GET\",\"payloadFormatVersion\":\"1.0\",\"type\":\"HTTP_PROXY\",\"uri\":\"https://ip-ranges.amazonaws.com/ip-ranges.json\"}}},\"/path1/path2\":{\"get\":{\"x-amazon-apigateway-integration\":{\"httpMethod\":\"GET\",\"payloadFormatVersion\":\"1.0\",\"type\":\"HTTP_PROXY\",\"uri\":\"https://ip-ranges.amazonaws.com/ip-ranges.json\"}}}}}",
					},
				},
				{
					Id:   "bar",
					Type: aws.AwsApiGatewayRestApiResourceType,
					Attrs: &resource.Attributes{
						"body": "{\"info\":{\"title\":\"example\",\"version\":\"1.0\"},\"openapi\":\"3.0.1\",\"paths\":{\"/path1\":{\"get\":{\"x-amazon-apigateway-integration\":{\"httpMethod\":\"GET\",\"payloadFormatVersion\":\"1.0\",\"type\":\"HTTP_PROXY\",\"uri\":\"https://ip-ranges.amazonaws.com/ip-ranges.json\"}}},\"/path1/path2\":{\"get\":{\"x-amazon-apigateway-integration\":{\"httpMethod\":\"GET\",\"payloadFormatVersion\":\"1.0\",\"type\":\"HTTP_PROXY\",\"uri\":\"https://ip-ranges.amazonaws.com/ip-ranges.json\"}}}}}",
					},
				},
			},
			remoteResources: []*resource.Resource{
				{
					Id:   "foo-path1",
					Type: aws.AwsApiGatewayResourceResourceType,
					Attrs: &resource.Attributes{
						"rest_api_id": "foo",
						"path":        "/path1",
					},
				},
				{
					Id:   "foo-path1-path2",
					Type: aws.AwsApiGatewayResourceResourceType,
					Attrs: &resource.Attributes{
						"rest_api_id": "foo",
						"path":        "/path1/path2",
					},
				},
				{
					Id:   "bar-path1",
					Type: aws.AwsApiGatewayResourceResourceType,
					Attrs: &resource.Attributes{
						"rest_api_id": "bar",
						"path":        "/path1",
					},
				},
				{
					Id:   "bar-path1-path2",
					Type: aws.AwsApiGatewayResourceResourceType,
					Attrs: &resource.Attributes{
						"rest_api_id": "bar",
						"path":        "/path1/path2",
					},
				},
			},
			expected: []*resource.Resource{
				{
					Id:   "foo",
					Type: aws.AwsApiGatewayRestApiResourceType,
					Attrs: &resource.Attributes{
						"body": "{\"info\":{\"title\":\"example\",\"version\":\"1.0\"},\"openapi\":\"3.0.1\",\"paths\":{\"/path1\":{\"get\":{\"x-amazon-apigateway-integration\":{\"httpMethod\":\"GET\",\"payloadFormatVersion\":\"1.0\",\"type\":\"HTTP_PROXY\",\"uri\":\"https://ip-ranges.amazonaws.com/ip-ranges.json\"}}},\"/path1/path2\":{\"get\":{\"x-amazon-apigateway-integration\":{\"httpMethod\":\"GET\",\"payloadFormatVersion\":\"1.0\",\"type\":\"HTTP_PROXY\",\"uri\":\"https://ip-ranges.amazonaws.com/ip-ranges.json\"}}}}}",
					},
				},
				{
					Id:   "foo-path1",
					Type: aws.AwsApiGatewayResourceResourceType,
					Attrs: &resource.Attributes{
						"rest_api_id": "foo",
						"path":        "/path1",
					},
				},
				{
					Id:   "foo-path1-path2",
					Type: aws.AwsApiGatewayResourceResourceType,
					Attrs: &resource.Attributes{
						"rest_api_id": "foo",
						"path":        "/path1/path2",
					},
				},
				{
					Id:   "bar",
					Type: aws.AwsApiGatewayRestApiResourceType,
					Attrs: &resource.Attributes{
						"body": "{\"info\":{\"title\":\"example\",\"version\":\"1.0\"},\"openapi\":\"3.0.1\",\"paths\":{\"/path1\":{\"get\":{\"x-amazon-apigateway-integration\":{\"httpMethod\":\"GET\",\"payloadFormatVersion\":\"1.0\",\"type\":\"HTTP_PROXY\",\"uri\":\"https://ip-ranges.amazonaws.com/ip-ranges.json\"}}},\"/path1/path2\":{\"get\":{\"x-amazon-apigateway-integration\":{\"httpMethod\":\"GET\",\"payloadFormatVersion\":\"1.0\",\"type\":\"HTTP_PROXY\",\"uri\":\"https://ip-ranges.amazonaws.com/ip-ranges.json\"}}}}}",
					},
				},
				{
					Id:   "bar-path1",
					Type: aws.AwsApiGatewayResourceResourceType,
					Attrs: &resource.Attributes{
						"rest_api_id": "bar",
						"path":        "/path1",
					},
				},
				{
					Id:   "bar-path1-path2",
					Type: aws.AwsApiGatewayResourceResourceType,
					Attrs: &resource.Attributes{
						"rest_api_id": "bar",
						"path":        "/path1/path2",
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

			m := NewAwsApiGatewayRestApiExpander(factory)
			err := m.Execute(&tt.remoteResources, &tt.resourcesFromState)
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
