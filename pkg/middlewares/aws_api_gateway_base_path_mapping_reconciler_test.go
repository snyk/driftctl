package middlewares

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/r3labs/diff/v2"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/aws"
)

func TestAwsApiGatewayBasePathMappingReconciler_Execute(t *testing.T) {
	tests := []struct {
		name               string
		resourcesFromState []*resource.Resource
		remoteResources    []*resource.Resource
		expected           []*resource.Resource
	}{
		{
			name: "with managed resources",
			resourcesFromState: []*resource.Resource{
				{
					Id:   "mapping1",
					Type: aws.AwsApiGatewayBasePathMappingResourceType,
				},
				{
					Id:   "mapping2",
					Type: aws.AwsApiGatewayV2MappingResourceType,
				},
			},
			remoteResources: []*resource.Resource{
				{
					Id:   "mapping1",
					Type: aws.AwsApiGatewayBasePathMappingResourceType,
				},
				{
					Id:   "mapping1",
					Type: aws.AwsApiGatewayV2MappingResourceType,
				},
				{
					Id:   "mapping2",
					Type: aws.AwsApiGatewayBasePathMappingResourceType,
				},
				{
					Id:   "mapping2",
					Type: aws.AwsApiGatewayV2MappingResourceType,
				},
			},
			expected: []*resource.Resource{
				{
					Id:   "mapping1",
					Type: aws.AwsApiGatewayBasePathMappingResourceType,
				},
				{
					Id:   "mapping2",
					Type: aws.AwsApiGatewayV2MappingResourceType,
				},
			},
		},
		{
			name:               "with unmanaged resources",
			resourcesFromState: []*resource.Resource{},
			remoteResources: []*resource.Resource{
				{
					Id:   "mapping1",
					Type: aws.AwsApiGatewayBasePathMappingResourceType,
				},
				{
					Id:   "mapping1",
					Type: aws.AwsApiGatewayV2MappingResourceType,
				},
				{
					Id:   "mapping2",
					Type: aws.AwsApiGatewayBasePathMappingResourceType,
				},
				{
					Id:   "mapping2",
					Type: aws.AwsApiGatewayV2MappingResourceType,
				},
			},
			expected: []*resource.Resource{
				{
					Id:   "mapping1",
					Type: aws.AwsApiGatewayBasePathMappingResourceType,
				},
				{
					Id:   "mapping2",
					Type: aws.AwsApiGatewayBasePathMappingResourceType,
				},
			},
		},
		{
			name: "with deleted resources",
			resourcesFromState: []*resource.Resource{
				{
					Id:   "mapping1",
					Type: aws.AwsApiGatewayBasePathMappingResourceType,
				},
				{
					Id:   "mapping2",
					Type: aws.AwsApiGatewayV2MappingResourceType,
				},
			},
			remoteResources: []*resource.Resource{},
			expected:        []*resource.Resource{},
		},
		{
			name: "with a mix of managed, unmanaged and deleted resources",
			resourcesFromState: []*resource.Resource{
				{
					Id:   "mapping1",
					Type: aws.AwsApiGatewayBasePathMappingResourceType,
				},
				{
					Id:   "mapping2",
					Type: aws.AwsApiGatewayV2MappingResourceType,
				},
				{
					Id:   "mapping4",
					Type: aws.AwsApiGatewayBasePathMappingResourceType,
				},
			},
			remoteResources: []*resource.Resource{
				{
					Id:   "mapping1",
					Type: aws.AwsApiGatewayBasePathMappingResourceType,
				},
				{
					Id:   "mapping1",
					Type: aws.AwsApiGatewayV2MappingResourceType,
				},
				{
					Id:   "mapping2",
					Type: aws.AwsApiGatewayBasePathMappingResourceType,
				},
				{
					Id:   "mapping2",
					Type: aws.AwsApiGatewayV2MappingResourceType,
				},
				{
					Id:   "mapping3",
					Type: aws.AwsApiGatewayBasePathMappingResourceType,
				},
				{
					Id:   "mapping3",
					Type: aws.AwsApiGatewayV2MappingResourceType,
				},
			},
			expected: []*resource.Resource{
				{
					Id:   "mapping1",
					Type: aws.AwsApiGatewayBasePathMappingResourceType,
				},
				{
					Id:   "mapping2",
					Type: aws.AwsApiGatewayV2MappingResourceType,
				},
				{
					Id:   "mapping3",
					Type: aws.AwsApiGatewayBasePathMappingResourceType,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewAwsApiGatewayBasePathMappingReconciler()
			err := m.Execute(&tt.remoteResources, &tt.resourcesFromState)
			if err != nil {
				t.Fatal(err)
			}
			changelog, err := diff.Diff(tt.expected, tt.remoteResources)
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
