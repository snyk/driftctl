package middlewares

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/r3labs/diff/v2"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/aws"
)

func TestAwsDefaultInternetGatewayRoute_Execute(t *testing.T) {
	tests := []struct {
		name               string
		remoteResources    []*resource.Resource
		resourcesFromState []*resource.Resource
		expected           []*resource.Resource
	}{
		{
			"default internet gateway route is not ignored when managed by IaC",
			[]*resource.Resource{
				{
					Id:   "default-vpc",
					Type: aws.AwsDefaultVpcResourceType,
				},
				{
					Id:   "default-igw",
					Type: aws.AwsInternetGatewayResourceType,
					Attrs: &resource.Attributes{
						"vpc_id": "default-vpc",
					},
				},
				{
					Id:   "default-route-table",
					Type: aws.AwsDefaultRouteTableResourceType,
					Attrs: &resource.Attributes{
						"vpc_id": "default-vpc",
					},
				},
				{
					Id:   "default-igw-route",
					Type: aws.AwsRouteResourceType,
					Attrs: &resource.Attributes{
						"route_table_id": "default-route-table",
						"gateway_id":     "default-igw",
					},
				},
				{
					Id:   "dummy-route",
					Type: aws.AwsRouteResourceType,
					Attrs: &resource.Attributes{
						"route_table_id": "default-route-table",
						"gateway_id":     "local",
					},
				},
			},
			[]*resource.Resource{
				{
					Id:   "default-igw-route",
					Type: aws.AwsRouteResourceType,
					Attrs: &resource.Attributes{
						"route_table_id": "default-route-table",
						"gateway_id":     "default-igw",
					},
				},
			},
			[]*resource.Resource{
				{
					Id:   "default-vpc",
					Type: aws.AwsDefaultVpcResourceType,
				},
				{
					Id:   "default-igw",
					Type: aws.AwsInternetGatewayResourceType,
					Attrs: &resource.Attributes{
						"vpc_id": "default-vpc",
					},
				},
				{
					Id:   "default-route-table",
					Type: aws.AwsDefaultRouteTableResourceType,
					Attrs: &resource.Attributes{
						"vpc_id": "default-vpc",
					},
				},
				{
					Id:   "default-igw-route",
					Type: aws.AwsRouteResourceType,
					Attrs: &resource.Attributes{
						"route_table_id": "default-route-table",
						"gateway_id":     "default-igw",
					},
				},
				{
					Id:   "dummy-route",
					Type: aws.AwsRouteResourceType,
					Attrs: &resource.Attributes{
						"route_table_id": "default-route-table",
						"gateway_id":     "local",
					},
				},
			},
		},
		{
			"default internet gateway route is ignored when not managed by IaC",
			[]*resource.Resource{
				{
					Id:   "default-vpc",
					Type: aws.AwsDefaultVpcResourceType,
				},
				{
					Id:   "default-igw",
					Type: aws.AwsInternetGatewayResourceType,
					Attrs: &resource.Attributes{
						"vpc_id": "default-vpc",
					},
				},
				{
					Id:   "default-route-table",
					Type: aws.AwsDefaultRouteTableResourceType,
					Attrs: &resource.Attributes{
						"vpc_id": "default-vpc",
					},
				},
				{
					Id:   "default-igw-route",
					Type: aws.AwsRouteResourceType,
					Attrs: &resource.Attributes{
						"route_table_id":         "default-route-table",
						"gateway_id":             "default-igw",
						"destination_cidr_block": "0.0.0.0/0",
					},
				},
				{
					Id:   "default-igw-non-default-route",
					Type: aws.AwsRouteResourceType,
					Attrs: &resource.Attributes{
						"route_table_id":         "default-route-table",
						"gateway_id":             "default-igw",
						"destination_cidr_block": "10.0.1.0/24",
					},
				},
				{
					Id:   "default-igw-default-ipv6-route",
					Type: aws.AwsRouteResourceType,
					Attrs: &resource.Attributes{
						"route_table_id":              "default-route-table",
						"gateway_id":                  "default-igw",
						"destination_ipv6_cidr_block": "::/0",
					},
				},
				{
					Id:   "dummy-route",
					Type: aws.AwsRouteResourceType,
					Attrs: &resource.Attributes{
						"route_table_id": "default-route-table",
						"gateway_id":     "local",
					},
				},
			},
			[]*resource.Resource{},
			[]*resource.Resource{
				{
					Id:   "default-vpc",
					Type: aws.AwsDefaultVpcResourceType,
				},
				{
					Id:   "default-igw",
					Type: aws.AwsInternetGatewayResourceType,
					Attrs: &resource.Attributes{
						"vpc_id": "default-vpc",
					},
				},
				{
					Id:   "default-route-table",
					Type: aws.AwsDefaultRouteTableResourceType,
					Attrs: &resource.Attributes{
						"vpc_id": "default-vpc",
					},
				},
				{
					Id:   "default-igw-non-default-route",
					Type: aws.AwsRouteResourceType,
					Attrs: &resource.Attributes{
						"route_table_id":         "default-route-table",
						"gateway_id":             "default-igw",
						"destination_cidr_block": "10.0.1.0/24",
					},
				},
				{
					Id:   "default-igw-default-ipv6-route",
					Type: aws.AwsRouteResourceType,
					Attrs: &resource.Attributes{
						"route_table_id":              "default-route-table",
						"gateway_id":                  "default-igw",
						"destination_ipv6_cidr_block": "::/0",
					},
				},
				{
					Id:   "dummy-route",
					Type: aws.AwsRouteResourceType,
					Attrs: &resource.Attributes{
						"route_table_id": "default-route-table",
						"gateway_id":     "local",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewAwsDefaultInternetGatewayRoute()
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
