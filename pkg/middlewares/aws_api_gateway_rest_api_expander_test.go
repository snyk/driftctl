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
			name: "create aws_api_gateway_resource from OpenAPI v3 JSON document",
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
				factory.On(
					"CreateAbstractResource",
					aws.AwsApiGatewayMethodResourceType,
					"agm-foo-bar-GET",
					map[string]interface{}{},
				).Once().Return(&resource.Resource{
					Id:    "agm-foo-bar-GET",
					Type:  aws.AwsApiGatewayMethodResourceType,
					Attrs: &resource.Attributes{},
				})
				factory.On(
					"CreateAbstractResource",
					aws.AwsApiGatewayMethodResourceType,
					"agm-foo-baz-GET",
					map[string]interface{}{},
				).Once().Return(&resource.Resource{
					Id:    "agm-foo-baz-GET",
					Type:  aws.AwsApiGatewayMethodResourceType,
					Attrs: &resource.Attributes{},
				})
				factory.On(
					"CreateAbstractResource",
					aws.AwsApiGatewayMethodResponseResourceType,
					"agmr-foo-bar-GET-200",
					map[string]interface{}{},
				).Once().Return(&resource.Resource{
					Id:    "agmr-foo-bar-GET-200",
					Type:  aws.AwsApiGatewayMethodResponseResourceType,
					Attrs: &resource.Attributes{},
				})
			},
			resourcesFromState: []*resource.Resource{
				{
					Id:   "foo",
					Type: aws.AwsApiGatewayRestApiResourceType,
					Attrs: &resource.Attributes{
						"body": "{\"info\":{\"title\":\"example\",\"version\":\"1.0\"},\"openapi\":\"3.0.1\",\"paths\":{\"/path1\":{\"get\":{\"parameters\":[{\"in\":\"query\",\"name\":\"type\",\"schema\":{\"type\":\"string\"}},{\"in\":\"query\",\"name\":\"page\",\"schema\":{\"type\":\"string\"}}],\"responses\":{\"200\":{\"content\":{\"application/json\":{\"schema\":{\"$ref\":\"#/components/schemas/Pets\"}}},\"description\":\"200 response\",\"headers\":{\"Access-Control-Allow-Origin\":{\"schema\":{\"type\":\"string\"}}}}},\"x-amazon-apigateway-integration\":{\"httpMethod\":\"GET\",\"payloadFormatVersion\":\"1.0\",\"type\":\"HTTP_PROXY\",\"uri\":\"https://ip-ranges.amazonaws.com/ip-ranges.json\"}}},\"/path1/path2\":{\"get\":{\"x-amazon-apigateway-integration\":{\"httpMethod\":\"GET\",\"payloadFormatVersion\":\"1.0\",\"type\":\"HTTP_PROXY\",\"uri\":\"https://ip-ranges.amazonaws.com/ip-ranges.json\"}}}}}",
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
				{
					Id:    "agm-foo-bar-GET",
					Type:  aws.AwsApiGatewayMethodResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:    "agm-foo-baz-GET",
					Type:  aws.AwsApiGatewayMethodResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:    "agmr-foo-bar-GET-200",
					Type:  aws.AwsApiGatewayMethodResponseResourceType,
					Attrs: &resource.Attributes{},
				},
			},
			expected: []*resource.Resource{
				{
					Id:   "foo",
					Type: aws.AwsApiGatewayRestApiResourceType,
					Attrs: &resource.Attributes{
						"body": "{\"info\":{\"title\":\"example\",\"version\":\"1.0\"},\"openapi\":\"3.0.1\",\"paths\":{\"/path1\":{\"get\":{\"parameters\":[{\"in\":\"query\",\"name\":\"type\",\"schema\":{\"type\":\"string\"}},{\"in\":\"query\",\"name\":\"page\",\"schema\":{\"type\":\"string\"}}],\"responses\":{\"200\":{\"content\":{\"application/json\":{\"schema\":{\"$ref\":\"#/components/schemas/Pets\"}}},\"description\":\"200 response\",\"headers\":{\"Access-Control-Allow-Origin\":{\"schema\":{\"type\":\"string\"}}}}},\"x-amazon-apigateway-integration\":{\"httpMethod\":\"GET\",\"payloadFormatVersion\":\"1.0\",\"type\":\"HTTP_PROXY\",\"uri\":\"https://ip-ranges.amazonaws.com/ip-ranges.json\"}}},\"/path1/path2\":{\"get\":{\"x-amazon-apigateway-integration\":{\"httpMethod\":\"GET\",\"payloadFormatVersion\":\"1.0\",\"type\":\"HTTP_PROXY\",\"uri\":\"https://ip-ranges.amazonaws.com/ip-ranges.json\"}}}}}",
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
				{
					Id:    "agm-foo-bar-GET",
					Type:  aws.AwsApiGatewayMethodResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:    "agm-foo-baz-GET",
					Type:  aws.AwsApiGatewayMethodResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:    "agmr-foo-bar-GET-200",
					Type:  aws.AwsApiGatewayMethodResponseResourceType,
					Attrs: &resource.Attributes{},
				},
			},
		},
		{
			name: "create aws_api_gateway_resource from OpenAPI v2 JSON document",
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
				factory.On(
					"CreateAbstractResource",
					aws.AwsApiGatewayMethodResourceType,
					"agm-foo-bar-GET",
					map[string]interface{}{},
				).Once().Return(&resource.Resource{
					Id:    "agm-foo-bar-GET",
					Type:  aws.AwsApiGatewayMethodResourceType,
					Attrs: &resource.Attributes{},
				})
				factory.On(
					"CreateAbstractResource",
					aws.AwsApiGatewayMethodResponseResourceType,
					"agmr-foo-bar-GET-200",
					map[string]interface{}{},
				).Once().Return(&resource.Resource{
					Id:    "agmr-foo-bar-GET-200",
					Type:  aws.AwsApiGatewayMethodResponseResourceType,
					Attrs: &resource.Attributes{},
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
				{
					Id:    "agm-foo-bar-GET",
					Type:  aws.AwsApiGatewayMethodResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:    "agmr-foo-bar-GET-200",
					Type:  aws.AwsApiGatewayMethodResponseResourceType,
					Attrs: &resource.Attributes{},
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
				{
					Id:    "agm-foo-bar-GET",
					Type:  aws.AwsApiGatewayMethodResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:    "agmr-foo-bar-GET-200",
					Type:  aws.AwsApiGatewayMethodResponseResourceType,
					Attrs: &resource.Attributes{},
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
			name: "unknown resource in body (e.g. missing resources)",
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
					Id:    "bar",
					Type:  aws.AwsApiGatewayRestApiResourceType,
					Attrs: &resource.Attributes{},
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
				factory.On(
					"CreateAbstractResource",
					aws.AwsApiGatewayMethodResourceType,
					"agm-foo-foo-path1-GET",
					map[string]interface{}{},
				).Once().Return(&resource.Resource{
					Id:    "agm-foo-foo-path1-GET",
					Type:  aws.AwsApiGatewayMethodResourceType,
					Attrs: &resource.Attributes{},
				})
				factory.On(
					"CreateAbstractResource",
					aws.AwsApiGatewayMethodResourceType,
					"agm-foo-foo-path1-path2-GET",
					map[string]interface{}{},
				).Once().Return(&resource.Resource{
					Id:    "agm-foo-foo-path1-path2-GET",
					Type:  aws.AwsApiGatewayMethodResourceType,
					Attrs: &resource.Attributes{},
				})
				factory.On(
					"CreateAbstractResource",
					aws.AwsApiGatewayMethodResourceType,
					"agm-bar-bar-path1-GET",
					map[string]interface{}{},
				).Once().Return(&resource.Resource{
					Id:    "agm-bar-bar-path1-GET",
					Type:  aws.AwsApiGatewayMethodResourceType,
					Attrs: &resource.Attributes{},
				})
				factory.On(
					"CreateAbstractResource",
					aws.AwsApiGatewayMethodResourceType,
					"agm-bar-bar-path1-path2-GET",
					map[string]interface{}{},
				).Once().Return(&resource.Resource{
					Id:    "agm-bar-bar-path1-path2-GET",
					Type:  aws.AwsApiGatewayMethodResourceType,
					Attrs: &resource.Attributes{},
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
				{
					Id:    "agm-foo-foo-path1-GET",
					Type:  aws.AwsApiGatewayMethodResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:    "agm-foo-foo-path1-path2-GET",
					Type:  aws.AwsApiGatewayMethodResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:    "agm-bar-bar-path1-GET",
					Type:  aws.AwsApiGatewayMethodResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:    "agm-bar-bar-path1-path2-GET",
					Type:  aws.AwsApiGatewayMethodResourceType,
					Attrs: &resource.Attributes{},
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
				{
					Id:    "agm-foo-foo-path1-GET",
					Type:  aws.AwsApiGatewayMethodResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:    "agm-foo-foo-path1-path2-GET",
					Type:  aws.AwsApiGatewayMethodResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:    "agm-bar-bar-path1-GET",
					Type:  aws.AwsApiGatewayMethodResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:    "agm-bar-bar-path1-path2-GET",
					Type:  aws.AwsApiGatewayMethodResourceType,
					Attrs: &resource.Attributes{},
				},
			},
		},
		{
			name: "create api gateway resources from OpenAPI v3 YAML document",
			mocks: func(factory *terraform.MockResourceFactory) {
				factory.On(
					"CreateAbstractResource",
					aws.AwsApiGatewayResourceResourceType,
					"bar",
					map[string]interface{}{
						"rest_api_id": "foo",
						"path":        "/{path+}",
					},
				).Once().Return(&resource.Resource{
					Id:   "bar",
					Type: aws.AwsApiGatewayResourceResourceType,
					Attrs: &resource.Attributes{
						"rest_api_id": "foo",
						"path":        "/{path+}",
					},
				})
				factory.On(
					"CreateAbstractResource",
					aws.AwsApiGatewayMethodResourceType,
					"agm-foo-bar-GET",
					map[string]interface{}{},
				).Once().Return(&resource.Resource{
					Id:    "agm-foo-bar-GET",
					Type:  aws.AwsApiGatewayMethodResourceType,
					Attrs: &resource.Attributes{},
				})
				factory.On(
					"CreateAbstractResource",
					aws.AwsApiGatewayMethodResponseResourceType,
					"agmr-foo-bar-GET-200",
					map[string]interface{}{},
				).Once().Return(&resource.Resource{
					Id:    "agmr-foo-bar-GET-200",
					Type:  aws.AwsApiGatewayMethodResponseResourceType,
					Attrs: &resource.Attributes{},
				})
			},
			resourcesFromState: []*resource.Resource{
				{
					Id:   "foo",
					Type: aws.AwsApiGatewayRestApiResourceType,
					Attrs: &resource.Attributes{
						"body": "---\nopenapi: \"3.0.1\"\ninfo:\n  title: baz\n  description: ComputelessBlog\npaths:\n  /{path+}:\n    get:\n      responses:\n        200:\n          description: \"200 response\"\n          content:\n            text/html:\n              schema:\n                $ref: \"#/components/schemas/Empty\"\n      x-amazon-apigateway-integration:\n        type: \"mock\"\n        responses:\n          default:\n            statusCode: \"200\"\n        passthroughBehavior: \"never\"\n        httpMethod: \"POST\"\ncomponents:\n  schemas:\n    Empty:\n      type: object\n      title: Empty Schema\n      description: Empty Schema",
					},
				},
			},
			remoteResources: []*resource.Resource{
				{
					Id:    "foo",
					Type:  aws.AwsApiGatewayRestApiResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:   "bar",
					Type: aws.AwsApiGatewayResourceResourceType,
					Attrs: &resource.Attributes{
						"rest_api_id": "foo",
						"path":        "/{path+}",
					},
				},
			},
			expected: []*resource.Resource{
				{
					Id:   "foo",
					Type: aws.AwsApiGatewayRestApiResourceType,
					Attrs: &resource.Attributes{
						"body": "---\nopenapi: \"3.0.1\"\ninfo:\n  title: baz\n  description: ComputelessBlog\npaths:\n  /{path+}:\n    get:\n      responses:\n        200:\n          description: \"200 response\"\n          content:\n            text/html:\n              schema:\n                $ref: \"#/components/schemas/Empty\"\n      x-amazon-apigateway-integration:\n        type: \"mock\"\n        responses:\n          default:\n            statusCode: \"200\"\n        passthroughBehavior: \"never\"\n        httpMethod: \"POST\"\ncomponents:\n  schemas:\n    Empty:\n      type: object\n      title: Empty Schema\n      description: Empty Schema",
					},
				},
				{
					Id:   "bar",
					Type: aws.AwsApiGatewayResourceResourceType,
					Attrs: &resource.Attributes{
						"rest_api_id": "foo",
						"path":        "/{path+}",
					},
				},
				{
					Id:    "agm-foo-bar-GET",
					Type:  aws.AwsApiGatewayMethodResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:    "agmr-foo-bar-GET-200",
					Type:  aws.AwsApiGatewayMethodResponseResourceType,
					Attrs: &resource.Attributes{},
				},
			},
		},
		{
			name: "create api gateway resources from OpenAPI v2 YAML document",
			mocks: func(factory *terraform.MockResourceFactory) {
				factory.On(
					"CreateAbstractResource",
					aws.AwsApiGatewayResourceResourceType,
					"bar",
					map[string]interface{}{
						"rest_api_id": "foo",
						"path":        "/webhook",
					},
				).Once().Return(&resource.Resource{
					Id:   "bar",
					Type: aws.AwsApiGatewayResourceResourceType,
					Attrs: &resource.Attributes{
						"rest_api_id": "foo",
						"path":        "/webhook",
					},
				})
				factory.On(
					"CreateAbstractResource",
					aws.AwsApiGatewayMethodResourceType,
					"agm-foo-bar-OPTIONS",
					map[string]interface{}{},
				).Once().Return(&resource.Resource{
					Id:    "agm-foo-bar-OPTIONS",
					Type:  aws.AwsApiGatewayMethodResourceType,
					Attrs: &resource.Attributes{},
				})
				factory.On(
					"CreateAbstractResource",
					aws.AwsApiGatewayMethodResponseResourceType,
					"agmr-foo-bar-OPTIONS-200",
					map[string]interface{}{},
				).Once().Return(&resource.Resource{
					Id:    "agmr-foo-bar-OPTIONS-200",
					Type:  aws.AwsApiGatewayMethodResponseResourceType,
					Attrs: &resource.Attributes{},
				})
			},
			resourcesFromState: []*resource.Resource{
				{
					Id:   "foo",
					Type: aws.AwsApiGatewayRestApiResourceType,
					Attrs: &resource.Attributes{
						"body": "---\nswagger: '2.0'\ninfo:\n  version: '1.0'\n  title: test\nschemes:\n  - https\npaths:\n  /webhook:\n    options:\n      consumes:\n        - application/json\n      produces:\n        - application/json\n      responses:\n        '200':\n          description: 200 response\n          schema:\n            $ref: \\\"#/definitions/Empty\\\"\n      x-amazon-apigateway-integration:\n        responses:\n          default:\n            statusCode: '200'\n        requestTemplates:\n          application/json: '{\\\"statusCode\\\": 200}'\n        passthroughBehavior: when_no_match\n        type: mock\n\n",
					},
				},
			},
			remoteResources: []*resource.Resource{
				{
					Id:    "foo",
					Type:  aws.AwsApiGatewayRestApiResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:   "bar",
					Type: aws.AwsApiGatewayResourceResourceType,
					Attrs: &resource.Attributes{
						"rest_api_id": "foo",
						"path":        "/webhook",
					},
				},
			},
			expected: []*resource.Resource{
				{
					Id:   "foo",
					Type: aws.AwsApiGatewayRestApiResourceType,
					Attrs: &resource.Attributes{
						"body": "---\nswagger: '2.0'\ninfo:\n  version: '1.0'\n  title: test\nschemes:\n  - https\npaths:\n  /webhook:\n    options:\n      consumes:\n        - application/json\n      produces:\n        - application/json\n      responses:\n        '200':\n          description: 200 response\n          schema:\n            $ref: \\\"#/definitions/Empty\\\"\n      x-amazon-apigateway-integration:\n        responses:\n          default:\n            statusCode: '200'\n        requestTemplates:\n          application/json: '{\\\"statusCode\\\": 200}'\n        passthroughBehavior: when_no_match\n        type: mock\n\n",
					},
				},
				{
					Id:   "bar",
					Type: aws.AwsApiGatewayResourceResourceType,
					Attrs: &resource.Attributes{
						"rest_api_id": "foo",
						"path":        "/webhook",
					},
				},
				{
					Id:    "agm-foo-bar-OPTIONS",
					Type:  aws.AwsApiGatewayMethodResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:    "agmr-foo-bar-OPTIONS-200",
					Type:  aws.AwsApiGatewayMethodResponseResourceType,
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
