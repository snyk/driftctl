package middlewares

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/r3labs/diff/v2"
	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

func TestAwsApiGatewayApiExpander_Execute(t *testing.T) {
	tests := []struct {
		name               string
		resourcesFromState []*resource.Resource
		remoteResources    []*resource.Resource
		mocks              func(*dctlresource.MockResourceFactory)
		expected           []*resource.Resource
	}{
		{
			name: "create aws_api_gateway_resource from OpenAPI v3 JSON document",
			mocks: func(factory *dctlresource.MockResourceFactory) {
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
				factory.On(
					"CreateAbstractResource",
					aws.AwsApiGatewayIntegrationResourceType,
					"agi-foo-bar-GET",
					map[string]interface{}{},
				).Once().Return(&resource.Resource{
					Id:    "agi-foo-bar-GET",
					Type:  aws.AwsApiGatewayIntegrationResourceType,
					Attrs: &resource.Attributes{},
				})
				factory.On(
					"CreateAbstractResource",
					aws.AwsApiGatewayIntegrationResourceType,
					"agi-foo-baz-GET",
					map[string]interface{}{},
				).Once().Return(&resource.Resource{
					Id:    "agi-foo-baz-GET",
					Type:  aws.AwsApiGatewayIntegrationResourceType,
					Attrs: &resource.Attributes{},
				})
				factory.On(
					"CreateAbstractResource",
					aws.AwsApiGatewayIntegrationResponseResourceType,
					"agir-foo-bar-GET-200",
					map[string]interface{}{},
				).Once().Return(&resource.Resource{
					Id:    "agir-foo-bar-GET-200",
					Type:  aws.AwsApiGatewayIntegrationResponseResourceType,
					Attrs: &resource.Attributes{},
				})
			},
			resourcesFromState: []*resource.Resource{
				{
					Id:   "foo",
					Type: aws.AwsApiGatewayRestApiResourceType,
					Attrs: &resource.Attributes{
						"body": "{\"info\":{\"title\":\"example\",\"version\":\"1.0\"},\"openapi\":\"3.0.1\",\"paths\":{\"/path1\":{\"get\":{\"parameters\":[{\"in\":\"query\",\"name\":\"type\",\"schema\":{\"type\":\"string\"}},{\"in\":\"query\",\"name\":\"page\",\"schema\":{\"type\":\"string\"}}],\"responses\":{\"200\":{\"content\":{\"application/json\":{\"schema\":{\"$ref\":\"#/components/schemas/Pets\"}}},\"description\":\"200 response\",\"headers\":{\"Access-Control-Allow-Origin\":{\"schema\":{\"type\":\"string\"}}}}},\"x-amazon-apigateway-integration\":{\"httpMethod\":\"GET\",\"payloadFormatVersion\":\"1.0\",\"type\":\"HTTP_PROXY\",\"uri\":\"https://ip-ranges.amazonaws.com/ip-ranges.json\",\"responses\":{\"2\\\\d{2}\":{\"responseTemplates\":{\"application/json\":\"#set ($root=$input.path('$')) { \\\"stage\\\": \\\"$root.name\\\", \\\"user-id\\\": \\\"$root.key\\\" }\",\"application/xml\":\"#set ($root=$input.path('$')) \\u003cstage\\u003e$root.name\\u003c/stage\\u003e \"},\"statusCode\":\"200\"}}}}},\"/path1/path2\":{\"get\":{\"x-amazon-apigateway-integration\":{\"httpMethod\":\"GET\",\"payloadFormatVersion\":\"1.0\",\"type\":\"HTTP_PROXY\",\"uri\":\"https://ip-ranges.amazonaws.com/ip-ranges.json\"}}}}}",
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
				{
					Id:    "agi-foo-bar-GET",
					Type:  aws.AwsApiGatewayIntegrationResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:    "agi-foo-baz-GET",
					Type:  aws.AwsApiGatewayIntegrationResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:    "agir-foo-bar-GET-200",
					Type:  aws.AwsApiGatewayIntegrationResponseResourceType,
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
				{
					Id:    "agi-foo-bar-GET",
					Type:  aws.AwsApiGatewayIntegrationResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:    "agi-foo-baz-GET",
					Type:  aws.AwsApiGatewayIntegrationResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:    "agir-foo-bar-GET-200",
					Type:  aws.AwsApiGatewayIntegrationResponseResourceType,
					Attrs: &resource.Attributes{},
				},
			},
		},
		{
			name: "create aws_api_gateway_resource from OpenAPI v2 JSON document",
			mocks: func(factory *dctlresource.MockResourceFactory) {
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
				factory.On(
					"CreateAbstractResource",
					aws.AwsApiGatewayIntegrationResourceType,
					"agi-foo-bar-GET",
					map[string]interface{}{},
				).Once().Return(&resource.Resource{
					Id:    "agi-foo-bar-GET",
					Type:  aws.AwsApiGatewayIntegrationResourceType,
					Attrs: &resource.Attributes{},
				})
				factory.On(
					"CreateAbstractResource",
					aws.AwsApiGatewayIntegrationResponseResourceType,
					"agir-foo-bar-GET-200",
					map[string]interface{}{},
				).Once().Return(&resource.Resource{
					Id:    "agir-foo-bar-GET-200",
					Type:  aws.AwsApiGatewayIntegrationResponseResourceType,
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
				{
					Id:    "agi-foo-bar-GET",
					Type:  aws.AwsApiGatewayIntegrationResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:    "agir-foo-bar-GET-200",
					Type:  aws.AwsApiGatewayIntegrationResponseResourceType,
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
				{
					Id:    "agi-foo-bar-GET",
					Type:  aws.AwsApiGatewayIntegrationResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:    "agir-foo-bar-GET-200",
					Type:  aws.AwsApiGatewayIntegrationResponseResourceType,
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
			mocks: func(factory *dctlresource.MockResourceFactory) {
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
				factory.On(
					"CreateAbstractResource",
					aws.AwsApiGatewayIntegrationResourceType,
					"agi-foo-foo-path1-GET",
					map[string]interface{}{},
				).Once().Return(&resource.Resource{
					Id:    "agi-foo-foo-path1-GET",
					Type:  aws.AwsApiGatewayIntegrationResourceType,
					Attrs: &resource.Attributes{},
				})
				factory.On(
					"CreateAbstractResource",
					aws.AwsApiGatewayIntegrationResourceType,
					"agi-foo-foo-path1-path2-GET",
					map[string]interface{}{},
				).Once().Return(&resource.Resource{
					Id:    "agi-foo-foo-path1-path2-GET",
					Type:  aws.AwsApiGatewayIntegrationResourceType,
					Attrs: &resource.Attributes{},
				})
				factory.On(
					"CreateAbstractResource",
					aws.AwsApiGatewayIntegrationResourceType,
					"agi-bar-bar-path1-GET",
					map[string]interface{}{},
				).Once().Return(&resource.Resource{
					Id:    "agi-bar-bar-path1-GET",
					Type:  aws.AwsApiGatewayIntegrationResourceType,
					Attrs: &resource.Attributes{},
				})
				factory.On(
					"CreateAbstractResource",
					aws.AwsApiGatewayIntegrationResourceType,
					"agi-bar-bar-path1-path2-GET",
					map[string]interface{}{},
				).Once().Return(&resource.Resource{
					Id:    "agi-bar-bar-path1-path2-GET",
					Type:  aws.AwsApiGatewayIntegrationResourceType,
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
				{
					Id:    "agi-foo-foo-path1-GET",
					Type:  aws.AwsApiGatewayIntegrationResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:    "agi-foo-foo-path1-path2-GET",
					Type:  aws.AwsApiGatewayIntegrationResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:    "agi-bar-bar-path1-GET",
					Type:  aws.AwsApiGatewayIntegrationResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:    "agi-bar-bar-path1-path2-GET",
					Type:  aws.AwsApiGatewayIntegrationResourceType,
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
				{
					Id:    "agi-foo-foo-path1-GET",
					Type:  aws.AwsApiGatewayIntegrationResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:    "agi-foo-foo-path1-path2-GET",
					Type:  aws.AwsApiGatewayIntegrationResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:    "agi-bar-bar-path1-GET",
					Type:  aws.AwsApiGatewayIntegrationResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:    "agi-bar-bar-path1-path2-GET",
					Type:  aws.AwsApiGatewayIntegrationResourceType,
					Attrs: &resource.Attributes{},
				},
			},
		},
		{
			name: "create gateway responses based on OpenAPI v2 and v3",
			mocks: func(factory *dctlresource.MockResourceFactory) {
				factory.On(
					"CreateAbstractResource",
					aws.AwsApiGatewayGatewayResponseResourceType,
					"aggr-v3-MISSING_AUTHENTICATION_TOKEN",
					map[string]interface{}{},
				).Once().Return(&resource.Resource{
					Id:    "aggr-v3-MISSING_AUTHENTICATION_TOKEN",
					Type:  aws.AwsApiGatewayGatewayResponseResourceType,
					Attrs: &resource.Attributes{},
				})
				factory.On(
					"CreateAbstractResource",
					aws.AwsApiGatewayGatewayResponseResourceType,
					"aggr-v2-MISSING_AUTHENTICATION_TOKEN",
					map[string]interface{}{},
				).Once().Return(&resource.Resource{
					Id:    "aggr-v2-MISSING_AUTHENTICATION_TOKEN",
					Type:  aws.AwsApiGatewayGatewayResponseResourceType,
					Attrs: &resource.Attributes{},
				})
			},
			resourcesFromState: []*resource.Resource{
				{
					Id:   "v3",
					Type: aws.AwsApiGatewayRestApiResourceType,
					Attrs: &resource.Attributes{
						"body": "{\"info\":{\"title\":\"example\",\"version\":\"1.0\"},\"openapi\":\"3.0.1\",\"paths\":{\"/path1\":{\"get\":{\"x-amazon-apigateway-integration\":{\"httpMethod\":\"GET\",\"payloadFormatVersion\":\"1.0\",\"type\":\"HTTP_PROXY\",\"uri\":\"https://ip-ranges.amazonaws.com/ip-ranges.json\"}}}},\"x-amazon-apigateway-gateway-responses\":{\"MISSING_AUTHENTICATION_TOKEN\":{\"responseParameters\":{\"gatewayresponse.header.Access-Control-Allow-Origin\":\"'a.b.c'\"},\"responseTemplates\":{\"application/json\":\"{\\n     \\\"message\\\": $context.error.messageString,\\n     \\\"type\\\":  \\\"$context.error.responseType\\\",\\n     \\\"stage\\\":  \\\"$context.stage\\\",\\n     \\\"resourcePath\\\":  \\\"$context.resourcePath\\\",\\n     \\\"stageVariables.a\\\":  \\\"$stageVariables.a\\\",\\n     \\\"statusCode\\\": \\\"'403'\\\"\\n}\"},\"statusCode\":403}}}",
					},
				},
				{
					Id:   "v2",
					Type: aws.AwsApiGatewayRestApiResourceType,
					Attrs: &resource.Attributes{
						"body": "{\"info\":{\"title\":\"test\",\"version\":\"2017-04-20T04:08:08Z\"},\"paths\":{\"/test\":{\"get\":{\"responses\":{\"200\":{\"description\":\"OK\"}},\"x-amazon-apigateway-integration\":{\"httpMethod\":\"GET\",\"responses\":{\"default\":{\"statusCode\":200}},\"type\":\"HTTP\",\"uri\":\"https://aws.amazon.com/\"}}}},\"schemes\":[\"https\"],\"swagger\":\"2.0\",\"x-amazon-apigateway-gateway-responses\":{\"MISSING_AUTHENTICATION_TOKEN\":{\"responseParameters\":{\"gatewayresponse.header.Access-Control-Allow-Origin\":\"'a.b.c'\"},\"responseTemplates\":{\"application/json\":\"{\\n     \\\"message\\\": $context.error.messageString,\\n     \\\"type\\\":  \\\"$context.error.responseType\\\",\\n     \\\"stage\\\":  \\\"$context.stage\\\",\\n     \\\"resourcePath\\\":  \\\"$context.resourcePath\\\",\\n     \\\"stageVariables.a\\\":  \\\"$stageVariables.a\\\",\\n     \\\"statusCode\\\": \\\"'403'\\\"\\n}\"},\"statusCode\":403}}}",
					},
				},
			},
			remoteResources: []*resource.Resource{},
			expected: []*resource.Resource{
				{
					Id:   "v3",
					Type: aws.AwsApiGatewayRestApiResourceType,
					Attrs: &resource.Attributes{
						"body": "{\"info\":{\"title\":\"example\",\"version\":\"1.0\"},\"openapi\":\"3.0.1\",\"paths\":{\"/path1\":{\"get\":{\"x-amazon-apigateway-integration\":{\"httpMethod\":\"GET\",\"payloadFormatVersion\":\"1.0\",\"type\":\"HTTP_PROXY\",\"uri\":\"https://ip-ranges.amazonaws.com/ip-ranges.json\"}}}},\"x-amazon-apigateway-gateway-responses\":{\"MISSING_AUTHENTICATION_TOKEN\":{\"responseParameters\":{\"gatewayresponse.header.Access-Control-Allow-Origin\":\"'a.b.c'\"},\"responseTemplates\":{\"application/json\":\"{\\n     \\\"message\\\": $context.error.messageString,\\n     \\\"type\\\":  \\\"$context.error.responseType\\\",\\n     \\\"stage\\\":  \\\"$context.stage\\\",\\n     \\\"resourcePath\\\":  \\\"$context.resourcePath\\\",\\n     \\\"stageVariables.a\\\":  \\\"$stageVariables.a\\\",\\n     \\\"statusCode\\\": \\\"'403'\\\"\\n}\"},\"statusCode\":403}}}",
					},
				},
				{
					Id:    "aggr-v3-MISSING_AUTHENTICATION_TOKEN",
					Type:  aws.AwsApiGatewayGatewayResponseResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:   "v2",
					Type: aws.AwsApiGatewayRestApiResourceType,
					Attrs: &resource.Attributes{
						"body": "{\"info\":{\"title\":\"test\",\"version\":\"2017-04-20T04:08:08Z\"},\"paths\":{\"/test\":{\"get\":{\"responses\":{\"200\":{\"description\":\"OK\"}},\"x-amazon-apigateway-integration\":{\"httpMethod\":\"GET\",\"responses\":{\"default\":{\"statusCode\":200}},\"type\":\"HTTP\",\"uri\":\"https://aws.amazon.com/\"}}}},\"schemes\":[\"https\"],\"swagger\":\"2.0\",\"x-amazon-apigateway-gateway-responses\":{\"MISSING_AUTHENTICATION_TOKEN\":{\"responseParameters\":{\"gatewayresponse.header.Access-Control-Allow-Origin\":\"'a.b.c'\"},\"responseTemplates\":{\"application/json\":\"{\\n     \\\"message\\\": $context.error.messageString,\\n     \\\"type\\\":  \\\"$context.error.responseType\\\",\\n     \\\"stage\\\":  \\\"$context.stage\\\",\\n     \\\"resourcePath\\\":  \\\"$context.resourcePath\\\",\\n     \\\"stageVariables.a\\\":  \\\"$stageVariables.a\\\",\\n     \\\"statusCode\\\": \\\"'403'\\\"\\n}\"},\"statusCode\":403}}}",
					},
				},
				{
					Id:    "aggr-v2-MISSING_AUTHENTICATION_TOKEN",
					Type:  aws.AwsApiGatewayGatewayResponseResourceType,
					Attrs: &resource.Attributes{},
				},
			},
		},
		{
			name: "create api gateway resources from OpenAPI v3 YAML document",
			mocks: func(factory *dctlresource.MockResourceFactory) {
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
				factory.On(
					"CreateAbstractResource",
					aws.AwsApiGatewayIntegrationResourceType,
					"agi-foo-bar-GET",
					map[string]interface{}{},
				).Once().Return(&resource.Resource{
					Id:    "agi-foo-bar-GET",
					Type:  aws.AwsApiGatewayIntegrationResourceType,
					Attrs: &resource.Attributes{},
				})
				factory.On(
					"CreateAbstractResource",
					aws.AwsApiGatewayIntegrationResponseResourceType,
					"agir-foo-bar-GET-200",
					map[string]interface{}{},
				).Once().Return(&resource.Resource{
					Id:    "agir-foo-bar-GET-200",
					Type:  aws.AwsApiGatewayIntegrationResponseResourceType,
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
				{
					Id:    "agi-foo-bar-GET",
					Type:  aws.AwsApiGatewayIntegrationResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:    "agir-foo-bar-GET-200",
					Type:  aws.AwsApiGatewayIntegrationResponseResourceType,
					Attrs: &resource.Attributes{},
				},
			},
		},
		{
			name: "create api gateway resources from OpenAPI v2 YAML document",
			mocks: func(factory *dctlresource.MockResourceFactory) {
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
				factory.On(
					"CreateAbstractResource",
					aws.AwsApiGatewayIntegrationResourceType,
					"agi-foo-bar-OPTIONS",
					map[string]interface{}{},
				).Once().Return(&resource.Resource{
					Id:    "agi-foo-bar-OPTIONS",
					Type:  aws.AwsApiGatewayIntegrationResourceType,
					Attrs: &resource.Attributes{},
				})
				factory.On(
					"CreateAbstractResource",
					aws.AwsApiGatewayIntegrationResponseResourceType,
					"agir-foo-bar-OPTIONS-200",
					map[string]interface{}{},
				).Once().Return(&resource.Resource{
					Id:    "agir-foo-bar-OPTIONS-200",
					Type:  aws.AwsApiGatewayIntegrationResponseResourceType,
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
				{
					Id:    "agi-foo-bar-OPTIONS",
					Type:  aws.AwsApiGatewayIntegrationResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:    "agir-foo-bar-OPTIONS-200",
					Type:  aws.AwsApiGatewayIntegrationResponseResourceType,
					Attrs: &resource.Attributes{},
				},
			},
		},
		{
			name: "creates routes from OpenAPI v3 YAML document (apigatewayv2)",
			resourcesFromState: []*resource.Resource{
				{
					Id:   "a-gateway",
					Type: aws.AwsApiGatewayV2ApiResourceType,
					Attrs: &resource.Attributes{
						"body": readFile(t, filepath.Join("testdata", "aws_apigatewayv2_api_body_openapiv3.yml")),
					},
				},
			},
			remoteResources: []*resource.Resource{
				{
					Id:   "openapi-derived-route-from-remote-1",
					Type: aws.AwsApiGatewayV2RouteResourceType,
					Attrs: &resource.Attributes{
						"api_id":    "a-gateway",
						"route_key": "GET /example",
					},
				},
				{
					Id:   "openapi-derived-route-from-remote-2",
					Type: aws.AwsApiGatewayV2RouteResourceType,
					Attrs: &resource.Attributes{
						"api_id":    "a-gateway",
						"route_key": "POST /example",
					},
				},
				{
					Id:   "openapi-derived-route-from-remote-3",
					Type: aws.AwsApiGatewayV2RouteResourceType,
					Attrs: &resource.Attributes{
						"api_id":    "a-gateway",
						"route_key": "GET /example2",
					},
				},
				{
					Id:   "irrelevant-route",
					Type: aws.AwsApiGatewayV2RouteResourceType,
					Attrs: &resource.Attributes{
						"api_id":    "another-gateway",
						"route_key": "GET /example2",
					},
				},
			},
			expected: []*resource.Resource{
				{
					Id:   "a-gateway",
					Type: aws.AwsApiGatewayV2ApiResourceType,
					Attrs: &resource.Attributes{
						"body": readFile(t, filepath.Join("testdata", "aws_apigatewayv2_api_body_openapiv3.yml")),
					},
				},
				{
					Id:   "openapi-derived-route-from-remote-1",
					Type: aws.AwsApiGatewayV2RouteResourceType,
				},
				{
					Id:   "openapi-derived-route-from-remote-2",
					Type: aws.AwsApiGatewayV2RouteResourceType,
				},
				{
					Id:   "openapi-derived-route-from-remote-3",
					Type: aws.AwsApiGatewayV2RouteResourceType,
				},
			},
			mocks: func(factory *dctlresource.MockResourceFactory) {
				factory.On("CreateAbstractResource", aws.AwsApiGatewayV2RouteResourceType, "openapi-derived-route-from-remote-1", map[string]interface{}{}).
					Once().Return(&resource.Resource{
					Id:   "openapi-derived-route-from-remote-1",
					Type: aws.AwsApiGatewayV2RouteResourceType,
				})
				factory.On("CreateAbstractResource", aws.AwsApiGatewayV2RouteResourceType, "openapi-derived-route-from-remote-2", map[string]interface{}{}).
					Once().Return(&resource.Resource{
					Id:   "openapi-derived-route-from-remote-2",
					Type: aws.AwsApiGatewayV2RouteResourceType,
				})
				factory.On("CreateAbstractResource", aws.AwsApiGatewayV2RouteResourceType, "openapi-derived-route-from-remote-3", map[string]interface{}{}).
					Once().Return(&resource.Resource{
					Id:   "openapi-derived-route-from-remote-3",
					Type: aws.AwsApiGatewayV2RouteResourceType,
				})
			},
		},
		{
			name: "creates routes from OpenAPI v2 JSON document (apigatewayv2)",
			resourcesFromState: []*resource.Resource{
				{
					Id:   "a-gateway",
					Type: aws.AwsApiGatewayV2ApiResourceType,
					Attrs: &resource.Attributes{
						"body": readFile(t, filepath.Join("testdata", "aws_apigatewayv2_api_body_openapiv2.json")),
					},
				},
			},
			remoteResources: []*resource.Resource{
				{
					Id:   "openapi-derived-route-from-remote",
					Type: aws.AwsApiGatewayV2RouteResourceType,
					Attrs: &resource.Attributes{
						"api_id":    "a-gateway",
						"route_key": "GET /example",
					},
				},
			},
			expected: []*resource.Resource{
				{
					Id:   "a-gateway",
					Type: aws.AwsApiGatewayV2ApiResourceType,
					Attrs: &resource.Attributes{
						"body": readFile(t, filepath.Join("testdata", "aws_apigatewayv2_api_body_openapiv2.json")),
					},
				},
				{
					Id:   "openapi-derived-route-from-remote",
					Type: aws.AwsApiGatewayV2RouteResourceType,
				},
			},
			mocks: func(factory *dctlresource.MockResourceFactory) {
				factory.On("CreateAbstractResource", aws.AwsApiGatewayV2RouteResourceType, "openapi-derived-route-from-remote", map[string]interface{}{}).
					Once().Return(&resource.Resource{
					Id:   "openapi-derived-route-from-remote",
					Type: aws.AwsApiGatewayV2RouteResourceType,
				})
			},
		},
		{
			name: "creates routes and integration from OpenAPI v2 JSON document (apigatewayv2)",
			resourcesFromState: []*resource.Resource{
				{
					Id:   "a-gateway",
					Type: aws.AwsApiGatewayV2ApiResourceType,
					Attrs: &resource.Attributes{
						"body": readFile(t, filepath.Join("testdata", "aws_apigatewayv2_api_body_integration_openapiv2.json")),
					},
				},
			},
			remoteResources: []*resource.Resource{
				{
					Id:   "openapi-derived-route-from-remote",
					Type: aws.AwsApiGatewayV2RouteResourceType,
					Attrs: &resource.Attributes{
						"api_id":    "a-gateway",
						"route_key": "GET /example",
					},
				},
				{
					Id:   "openapi-derived-integration-from-remote",
					Type: aws.AwsApiGatewayV2IntegrationResourceType,
					Attrs: &resource.Attributes{
						"api_id":             "a-gateway",
						"integration_type":   "HTTP_PROXY",
						"integration_method": "GET",
					},
				},
			},
			expected: []*resource.Resource{
				{
					Id:   "a-gateway",
					Type: aws.AwsApiGatewayV2ApiResourceType,
					Attrs: &resource.Attributes{
						"body": readFile(t, filepath.Join("testdata", "aws_apigatewayv2_api_body_integration_openapiv2.json")),
					},
				},
				{
					Id:   "openapi-derived-route-from-remote",
					Type: aws.AwsApiGatewayV2RouteResourceType,
				},
				{
					Id:   "openapi-derived-integration-from-remote",
					Type: aws.AwsApiGatewayV2IntegrationResourceType,
				},
			},
			mocks: func(factory *dctlresource.MockResourceFactory) {
				factory.On("CreateAbstractResource", aws.AwsApiGatewayV2RouteResourceType, "openapi-derived-route-from-remote", map[string]interface{}{}).
					Once().Return(&resource.Resource{
					Id:   "openapi-derived-route-from-remote",
					Type: aws.AwsApiGatewayV2RouteResourceType,
				})
				factory.On("CreateAbstractResource", aws.AwsApiGatewayV2IntegrationResourceType, "openapi-derived-integration-from-remote", map[string]interface{}{}).
					Once().Return(&resource.Resource{
					Id:   "openapi-derived-integration-from-remote",
					Type: aws.AwsApiGatewayV2IntegrationResourceType,
				})
			},
		},
		{
			name: "creates routes and integrations from OpenAPI v3 YAML document (apigatewayv2)",
			resourcesFromState: []*resource.Resource{
				{
					Id:   "a-gateway",
					Type: aws.AwsApiGatewayV2ApiResourceType,
					Attrs: &resource.Attributes{
						"body": readFile(t, filepath.Join("testdata", "aws_apigatewayv2_api_body_integration_openapiv3.yml")),
					},
				},
			},
			remoteResources: []*resource.Resource{
				{
					Id:   "openapi-derived-route-from-remote",
					Type: aws.AwsApiGatewayV2RouteResourceType,
					Attrs: &resource.Attributes{
						"api_id":    "a-gateway",
						"route_key": "GET /example",
					},
				},
				{
					Id:   "openapi-derived-integration-from-remote",
					Type: aws.AwsApiGatewayV2IntegrationResourceType,
					Attrs: &resource.Attributes{
						"api_id":             "a-gateway",
						"integration_type":   "HTTP_PROXY",
						"integration_method": "GET",
					},
				},
			},
			expected: []*resource.Resource{
				{
					Id:   "a-gateway",
					Type: aws.AwsApiGatewayV2ApiResourceType,
					Attrs: &resource.Attributes{
						"body": readFile(t, filepath.Join("testdata", "aws_apigatewayv2_api_body_integration_openapiv3.yml")),
					},
				},
				{
					Id:   "openapi-derived-route-from-remote",
					Type: aws.AwsApiGatewayV2RouteResourceType,
				},
				{
					Id:   "openapi-derived-integration-from-remote",
					Type: aws.AwsApiGatewayV2IntegrationResourceType,
				},
			},
			mocks: func(factory *dctlresource.MockResourceFactory) {
				factory.On("CreateAbstractResource", aws.AwsApiGatewayV2RouteResourceType, "openapi-derived-route-from-remote", map[string]interface{}{}).
					Once().Return(&resource.Resource{
					Id:   "openapi-derived-route-from-remote",
					Type: aws.AwsApiGatewayV2RouteResourceType,
				})
				factory.On("CreateAbstractResource", aws.AwsApiGatewayV2IntegrationResourceType, "openapi-derived-integration-from-remote", map[string]interface{}{}).
					Once().Return(&resource.Resource{
					Id:   "openapi-derived-integration-from-remote",
					Type: aws.AwsApiGatewayV2IntegrationResourceType,
				})
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := &dctlresource.MockResourceFactory{}
			if tt.mocks != nil {
				tt.mocks(factory)
			}

			m := NewAwsApiGatewayApiExpander(factory)
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
