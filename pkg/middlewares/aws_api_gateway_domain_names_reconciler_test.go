package middlewares

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/r3labs/diff/v2"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

func TestAwsApiGatewayDomainNamesReconciler_Execute(t *testing.T) {
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
					Id:   "domain1",
					Type: aws.AwsApiGatewayDomainNameResourceType,
				},
				{
					Id:   "domain2",
					Type: aws.AwsApiGatewayV2DomainNameResourceType,
				},
			},
			remoteResources: []*resource.Resource{
				{
					Id:   "domain1",
					Type: aws.AwsApiGatewayDomainNameResourceType,
				},
				{
					Id:   "domain1",
					Type: aws.AwsApiGatewayV2DomainNameResourceType,
				},
				{
					Id:   "domain2",
					Type: aws.AwsApiGatewayDomainNameResourceType,
				},
				{
					Id:   "domain2",
					Type: aws.AwsApiGatewayV2DomainNameResourceType,
				},
			},
			expected: []*resource.Resource{
				{
					Id:   "domain1",
					Type: aws.AwsApiGatewayDomainNameResourceType,
				},
				{
					Id:   "domain2",
					Type: aws.AwsApiGatewayV2DomainNameResourceType,
				},
			},
		},
		{
			name:               "with unmanaged resources",
			resourcesFromState: []*resource.Resource{},
			remoteResources: []*resource.Resource{
				{
					Id:   "domain1",
					Type: aws.AwsApiGatewayDomainNameResourceType,
				},
				{
					Id:   "domain1",
					Type: aws.AwsApiGatewayV2DomainNameResourceType,
				},
				{
					Id:   "domain2",
					Type: aws.AwsApiGatewayDomainNameResourceType,
				},
				{
					Id:   "domain2",
					Type: aws.AwsApiGatewayV2DomainNameResourceType,
				},
			},
			expected: []*resource.Resource{
				{
					Id:   "domain1",
					Type: aws.AwsApiGatewayDomainNameResourceType,
				},
				{
					Id:   "domain2",
					Type: aws.AwsApiGatewayDomainNameResourceType,
				},
			},
		},
		{
			name: "with deleted resources",
			resourcesFromState: []*resource.Resource{
				{
					Id:   "domain1",
					Type: aws.AwsApiGatewayDomainNameResourceType,
				},
				{
					Id:   "domain2",
					Type: aws.AwsApiGatewayV2DomainNameResourceType,
				},
			},
			remoteResources: []*resource.Resource{},
			expected:        []*resource.Resource{},
		},
		{
			name: "with a mix of managed, unmanaged and deleted resources",
			resourcesFromState: []*resource.Resource{
				{
					Id:   "domain1",
					Type: aws.AwsApiGatewayDomainNameResourceType,
				},
				{
					Id:   "domain2",
					Type: aws.AwsApiGatewayV2DomainNameResourceType,
				},
				{
					Id:   "domain4",
					Type: aws.AwsApiGatewayDomainNameResourceType,
				},
			},
			remoteResources: []*resource.Resource{
				{
					Id:   "domain1",
					Type: aws.AwsApiGatewayDomainNameResourceType,
				},
				{
					Id:   "domain1",
					Type: aws.AwsApiGatewayV2DomainNameResourceType,
				},
				{
					Id:   "domain2",
					Type: aws.AwsApiGatewayDomainNameResourceType,
				},
				{
					Id:   "domain2",
					Type: aws.AwsApiGatewayV2DomainNameResourceType,
				},
				{
					Id:   "domain3",
					Type: aws.AwsApiGatewayDomainNameResourceType,
				},
				{
					Id:   "domain3",
					Type: aws.AwsApiGatewayV2DomainNameResourceType,
				},
			},
			expected: []*resource.Resource{
				{
					Id:   "domain1",
					Type: aws.AwsApiGatewayDomainNameResourceType,
				},
				{
					Id:   "domain2",
					Type: aws.AwsApiGatewayV2DomainNameResourceType,
				},
				{
					Id:   "domain3",
					Type: aws.AwsApiGatewayDomainNameResourceType,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewAwsApiGatewayDomainNamesReconciler()
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
