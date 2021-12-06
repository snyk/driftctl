package middlewares

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/r3labs/diff/v2"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

func TestAwsDefaultRouteTable_Execute(t *testing.T) {
	tests := []struct {
		name               string
		remoteResources    []*resource.Resource
		resourcesFromState []*resource.Resource
		expected           []*resource.Resource
	}{
		{
			"test that default route tables are not excluded when managed by IaC",
			[]*resource.Resource{
				{
					Id:   "non-default-route-table",
					Type: aws.AwsRouteTableResourceType,
				},
				{
					Id:   "default-route-table",
					Type: aws.AwsDefaultRouteTableResourceType,
				},
			},
			[]*resource.Resource{
				{
					Id:   "non-default-route-table",
					Type: aws.AwsRouteTableResourceType,
				},
				{
					Id:   "default-route-table",
					Type: aws.AwsDefaultRouteTableResourceType,
				},
			},
			[]*resource.Resource{
				{
					Id:   "non-default-route-table",
					Type: aws.AwsRouteTableResourceType,
				},
				{
					Id:   "default-route-table",
					Type: aws.AwsDefaultRouteTableResourceType,
				},
			},
		},
		{
			"test that default route tables are excluded when not managed by IaC",
			[]*resource.Resource{
				{
					Id:   "non-default-route-table",
					Type: aws.AwsRouteTableResourceType,
				},
				{
					Id:   "default-route-table",
					Type: aws.AwsDefaultRouteTableResourceType,
				},
			},
			[]*resource.Resource{
				{
					Id:   "non-default-route-table",
					Type: aws.AwsRouteTableResourceType,
				},
			},
			[]*resource.Resource{
				{
					Id:   "non-default-route-table",
					Type: aws.AwsRouteTableResourceType,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewAwsDefaultRouteTable()
			err := m.Execute(&tt.remoteResources, &tt.resourcesFromState)
			if err != nil {
				t.Fatal(err)
			}

			changelog, err := diff.Diff(tt.remoteResources, tt.expected)
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
