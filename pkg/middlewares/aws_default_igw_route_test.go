package middlewares

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/r3labs/diff/v2"
)

func TestAwsDefaultInternetGatewayRoute_Execute(t *testing.T) {
	tests := []struct {
		name               string
		remoteResources    []resource.Resource
		resourcesFromState []resource.Resource
		expected           []resource.Resource
	}{
		{
			"default internet gateway route is not ignored when managed by IaC",
			[]resource.Resource{
				&resource.AbstractResource{
					Id:   "default-vpc",
					Type: aws.AwsDefaultVpcResourceType,
				},
				&resource.AbstractResource{
					Id:   "default-igw",
					Type: aws.AwsInternetGatewayResourceType,
					Attrs: &resource.Attributes{
						"vpc_id": "default-vpc",
					},
				},
				&resource.AbstractResource{
					Id:   "default-route-table",
					Type: aws.AwsDefaultRouteTableResourceType,
					Attrs: &resource.Attributes{
						"vpc_id": "default-vpc",
					},
				},
				&resource.AbstractResource{
					Id:   "default-igw-route",
					Type: aws.AwsRouteResourceType,
					Attrs: &resource.Attributes{
						"route_table_id": "default-route-table",
						"gateway_id":     "default-igw",
					},
				},
				&resource.AbstractResource{
					Id:   "dummy-route",
					Type: aws.AwsRouteResourceType,
					Attrs: &resource.Attributes{
						"route_table_id": "default-route-table",
						"gateway_id":     "local",
					},
				},
			},
			[]resource.Resource{
				&resource.AbstractResource{
					Id:   "default-igw-route",
					Type: aws.AwsRouteResourceType,
					Attrs: &resource.Attributes{
						"route_table_id": "default-route-table",
						"gateway_id":     "default-igw",
					},
				},
			},
			[]resource.Resource{
				&resource.AbstractResource{
					Id:   "default-vpc",
					Type: aws.AwsDefaultVpcResourceType,
				},
				&resource.AbstractResource{
					Id:   "default-igw",
					Type: aws.AwsInternetGatewayResourceType,
					Attrs: &resource.Attributes{
						"vpc_id": "default-vpc",
					},
				},
				&resource.AbstractResource{
					Id:   "default-route-table",
					Type: aws.AwsDefaultRouteTableResourceType,
					Attrs: &resource.Attributes{
						"vpc_id": "default-vpc",
					},
				},
				&resource.AbstractResource{
					Id:   "default-igw-route",
					Type: aws.AwsRouteResourceType,
					Attrs: &resource.Attributes{
						"route_table_id": "default-route-table",
						"gateway_id":     "default-igw",
					},
				},
				&resource.AbstractResource{
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
			[]resource.Resource{
				&resource.AbstractResource{
					Id:   "default-vpc",
					Type: aws.AwsDefaultVpcResourceType,
				},
				&resource.AbstractResource{
					Id:   "default-igw",
					Type: aws.AwsInternetGatewayResourceType,
					Attrs: &resource.Attributes{
						"vpc_id": "default-vpc",
					},
				},
				&resource.AbstractResource{
					Id:   "default-route-table",
					Type: aws.AwsDefaultRouteTableResourceType,
					Attrs: &resource.Attributes{
						"vpc_id": "default-vpc",
					},
				},
				&resource.AbstractResource{
					Id:   "default-igw-route",
					Type: aws.AwsRouteResourceType,
					Attrs: &resource.Attributes{
						"route_table_id":         "default-route-table",
						"gateway_id":             "default-igw",
						"destination_cidr_block": "0.0.0.0/0",
					},
				},
				&resource.AbstractResource{
					Id:   "default-igw-non-default-route",
					Type: aws.AwsRouteResourceType,
					Attrs: &resource.Attributes{
						"route_table_id":         "default-route-table",
						"gateway_id":             "default-igw",
						"destination_cidr_block": "10.0.1.0/24",
					},
				},
				&resource.AbstractResource{
					Id:   "default-igw-default-ipv6-route",
					Type: aws.AwsRouteResourceType,
					Attrs: &resource.Attributes{
						"route_table_id":              "default-route-table",
						"gateway_id":                  "default-igw",
						"destination_ipv6_cidr_block": "::/0",
					},
				},
				&resource.AbstractResource{
					Id:   "dummy-route",
					Type: aws.AwsRouteResourceType,
					Attrs: &resource.Attributes{
						"route_table_id": "default-route-table",
						"gateway_id":     "local",
					},
				},
			},
			[]resource.Resource{},
			[]resource.Resource{
				&resource.AbstractResource{
					Id:   "default-vpc",
					Type: aws.AwsDefaultVpcResourceType,
				},
				&resource.AbstractResource{
					Id:   "default-igw",
					Type: aws.AwsInternetGatewayResourceType,
					Attrs: &resource.Attributes{
						"vpc_id": "default-vpc",
					},
				},
				&resource.AbstractResource{
					Id:   "default-route-table",
					Type: aws.AwsDefaultRouteTableResourceType,
					Attrs: &resource.Attributes{
						"vpc_id": "default-vpc",
					},
				},
				&resource.AbstractResource{
					Id:   "default-igw-non-default-route",
					Type: aws.AwsRouteResourceType,
					Attrs: &resource.Attributes{
						"route_table_id":         "default-route-table",
						"gateway_id":             "default-igw",
						"destination_cidr_block": "10.0.1.0/24",
					},
				},
				&resource.AbstractResource{
					Id:   "default-igw-default-ipv6-route",
					Type: aws.AwsRouteResourceType,
					Attrs: &resource.Attributes{
						"route_table_id":              "default-route-table",
						"gateway_id":                  "default-igw",
						"destination_ipv6_cidr_block": "::/0",
					},
				},
				&resource.AbstractResource{
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
