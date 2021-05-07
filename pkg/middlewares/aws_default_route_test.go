package middlewares

import (
	"strings"
	"testing"

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
				&resource.AbstractResource{
					Id:   "a-dummy-route",
					Type: aws.AwsRouteResourceType,
					Attrs: &resource.Attributes{
						"route_table_id": "fake-table-id",
						"origin":         "CreateRoute",
					},
				},
				&resource.AbstractResource{
					Id:   "default-managed-by-IaC",
					Type: aws.AwsRouteResourceType,
					Attrs: &resource.Attributes{
						"route_table_id": "fake-table-id",
						"origin":         "CreateRouteTable",
					},
				},
			},
			[]resource.Resource{
				&resource.AbstractResource{
					Id:   "default-managed-by-IaC",
					Type: aws.AwsRouteResourceType,
					Attrs: &resource.Attributes{
						"route_table_id": "fake-table-id",
						"origin":         "CreateRouteTable",
					},
				},
			},
			[]resource.Resource{
				&resource2.FakeResource{
					Id: "fake",
				},
				&resource.AbstractResource{
					Id:   "a-dummy-route",
					Type: aws.AwsRouteResourceType,
					Attrs: &resource.Attributes{
						"route_table_id": "fake-table-id",
						"origin":         "CreateRoute",
					},
				},
				&resource.AbstractResource{
					Id:   "default-managed-by-IaC",
					Type: aws.AwsRouteResourceType,
					Attrs: &resource.Attributes{
						"route_table_id": "fake-table-id",
						"origin":         "CreateRouteTable",
					},
				},
			},
		},
		{
			"test that default routes are ignored when not managed by IaC",
			[]resource.Resource{
				&resource2.FakeResource{
					Id: "fake",
				},
				&resource.AbstractResource{
					Id:   "a-dummy-route",
					Type: aws.AwsRouteResourceType,
					Attrs: &resource.Attributes{
						"route_table_id": "fake-table-id",
						"origin":         "CreateRoute",
					},
				},
				&resource.AbstractResource{
					Id:   "default-managed-by-IaC",
					Type: aws.AwsRouteResourceType,
					Attrs: &resource.Attributes{
						"route_table_id": "fake-table-id",
						"origin":         "CreateRouteTable",
					},
				},
			},
			[]resource.Resource{},
			[]resource.Resource{
				&resource2.FakeResource{
					Id: "fake",
				},
				&resource.AbstractResource{
					Id:   "a-dummy-route",
					Type: aws.AwsRouteResourceType,
					Attrs: &resource.Attributes{
						"route_table_id": "fake-table-id",
						"origin":         "CreateRoute",
					},
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
