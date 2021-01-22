package middlewares

import (
	"strings"
	"testing"

	awssdk "github.com/aws/aws-sdk-go/aws"
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
				&aws.AwsDefaultVpc{
					Id: "default-vpc",
				},
				&aws.AwsInternetGateway{
					Id:    "default-igw",
					VpcId: awssdk.String("default-vpc"),
				},
				&aws.AwsDefaultRouteTable{
					Id:    "default-route-table",
					VpcId: awssdk.String("default-vpc"),
				},
				&aws.AwsRoute{
					Id:           "default-igw-route",
					RouteTableId: awssdk.String("default-route-table"),
					GatewayId:    awssdk.String("default-igw"),
				},
				&aws.AwsRoute{
					Id:           "dummy-route",
					RouteTableId: awssdk.String("default-route-table"),
					GatewayId:    awssdk.String("local"),
				},
			},
			[]resource.Resource{
				&aws.AwsRoute{
					Id:           "default-igw-route",
					RouteTableId: awssdk.String("default-route-table"),
					GatewayId:    awssdk.String("default-igw"),
				},
			},
			[]resource.Resource{
				&aws.AwsDefaultVpc{
					Id: "default-vpc",
				},
				&aws.AwsInternetGateway{
					Id:    "default-igw",
					VpcId: awssdk.String("default-vpc"),
				},
				&aws.AwsDefaultRouteTable{
					Id:    "default-route-table",
					VpcId: awssdk.String("default-vpc"),
				},
				&aws.AwsRoute{
					Id:           "default-igw-route",
					RouteTableId: awssdk.String("default-route-table"),
					GatewayId:    awssdk.String("default-igw"),
				},
				&aws.AwsRoute{
					Id:           "dummy-route",
					RouteTableId: awssdk.String("default-route-table"),
					GatewayId:    awssdk.String("local"),
				},
			},
		},
		{
			"default internet gateway route is ignored when not managed by IaC",
			[]resource.Resource{
				&aws.AwsDefaultVpc{
					Id: "default-vpc",
				},
				&aws.AwsInternetGateway{
					Id:    "default-igw",
					VpcId: awssdk.String("default-vpc"),
				},
				&aws.AwsDefaultRouteTable{
					Id:    "default-route-table",
					VpcId: awssdk.String("default-vpc"),
				},
				&aws.AwsRoute{
					Id:           "default-igw-route",
					RouteTableId: awssdk.String("default-route-table"),
					GatewayId:    awssdk.String("default-igw"),
				},
				&aws.AwsRoute{
					Id:           "dummy-route",
					RouteTableId: awssdk.String("default-route-table"),
					GatewayId:    awssdk.String("local"),
				},
			},
			[]resource.Resource{},
			[]resource.Resource{
				&aws.AwsDefaultVpc{
					Id: "default-vpc",
				},
				&aws.AwsInternetGateway{
					Id:    "default-igw",
					VpcId: awssdk.String("default-vpc"),
				},
				&aws.AwsDefaultRouteTable{
					Id:    "default-route-table",
					VpcId: awssdk.String("default-vpc"),
				},
				&aws.AwsRoute{
					Id:           "dummy-route",
					RouteTableId: awssdk.String("default-route-table"),
					GatewayId:    awssdk.String("local"),
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
