package middlewares

import (
	"strings"
	"testing"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	resource2 "github.com/cloudskiff/driftctl/test/resource"
	"github.com/r3labs/diff/v2"
)

func TestAwsDefaultRoute_Execute(t *testing.T) {

	tests := []struct {
		name               string
		remoteResources    []resource.Resource
		resourcesFromState []resource.Resource
		expected           []resource.Resource
	}{
		{
			"test that default routes are not ignored when managed by IaC",
			[]resource.Resource{
				&resource2.FakeResource{
					Id: "fake",
				},
				&aws.AwsRoute{
					Id:           "a-dummy-route",
					Origin:       awssdk.String("CreateRoute"),
					RouteTableId: awssdk.String("fake-table-id"),
				},
				&aws.AwsRoute{
					Id:           "default-managed-by-IaC",
					Origin:       awssdk.String("CreateRouteTable"),
					RouteTableId: awssdk.String("fake-table-id"),
				},
			},
			[]resource.Resource{
				&aws.AwsRoute{
					Id:           "default-managed-by-IaC",
					Origin:       awssdk.String("CreateRouteTable"),
					RouteTableId: awssdk.String("fake-table-id"),
				},
			},
			[]resource.Resource{
				&resource2.FakeResource{
					Id: "fake",
				},
				&aws.AwsRoute{
					Id:           "a-dummy-route",
					Origin:       awssdk.String("CreateRoute"),
					RouteTableId: awssdk.String("fake-table-id"),
				},
				&aws.AwsRoute{
					Id:           "default-managed-by-IaC",
					Origin:       awssdk.String("CreateRouteTable"),
					RouteTableId: awssdk.String("fake-table-id"),
				},
			},
		},
		{
			"test that default routes are ignored when not managed by IaC",
			[]resource.Resource{
				&resource2.FakeResource{
					Id: "fake",
				},
				&aws.AwsRoute{
					Id:           "a-dummy-route",
					Origin:       awssdk.String("CreateRoute"),
					RouteTableId: awssdk.String("fake-table-id"),
				},
				&aws.AwsRoute{
					Id:           "default-managed-by-IaC",
					Origin:       awssdk.String("CreateRouteTable"),
					RouteTableId: awssdk.String("fake-table-id"),
				},
			},
			[]resource.Resource{},
			[]resource.Resource{
				&resource2.FakeResource{
					Id: "fake",
				},
				&aws.AwsRoute{
					Id:           "a-dummy-route",
					Origin:       awssdk.String("CreateRoute"),
					RouteTableId: awssdk.String("fake-table-id"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewAwsDefaultRoute()
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
