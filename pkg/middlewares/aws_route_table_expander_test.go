package middlewares

import (
	"strings"
	"testing"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/cloudskiff/driftctl/mocks"
	"github.com/cloudskiff/driftctl/pkg/alerter"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	resource2 "github.com/cloudskiff/driftctl/test/resource"
	"github.com/r3labs/diff/v2"
)

func TestAwsRouteTableExpander_Execute(t *testing.T) {
	tests := []struct {
		name     string
		input    []resource.Resource
		expected []resource.Resource
	}{
		{
			"test with nil route attributes",
			[]resource.Resource{
				&aws.AwsRouteTable{
					Id:    "table_from_state",
					Route: nil,
				},
			},

			[]resource.Resource{
				&aws.AwsRouteTable{
					Id:    "table_from_state",
					Route: nil,
				},
			},
		},
		{
			"test with empty route attributes",
			[]resource.Resource{
				&aws.AwsRouteTable{
					Id: "table_from_state",
					Route: &[]struct {
						CidrBlock              *string `cty:"cidr_block"`
						EgressOnlyGatewayId    *string `cty:"egress_only_gateway_id"`
						GatewayId              *string `cty:"gateway_id"`
						InstanceId             *string `cty:"instance_id"`
						Ipv6CidrBlock          *string `cty:"ipv6_cidr_block"`
						LocalGatewayId         *string `cty:"local_gateway_id"`
						NatGatewayId           *string `cty:"nat_gateway_id"`
						NetworkInterfaceId     *string `cty:"network_interface_id"`
						TransitGatewayId       *string `cty:"transit_gateway_id"`
						VpcEndpointId          *string `cty:"vpc_endpoint_id"`
						VpcPeeringConnectionId *string `cty:"vpc_peering_connection_id"`
					}{},
				},
			},
			[]resource.Resource{
				&aws.AwsRouteTable{
					Id:    "table_from_state",
					Route: nil,
				},
			},
		},
		{
			"test route are expanded",
			[]resource.Resource{
				&resource2.FakeResource{
					Id: "fake_resource",
				},
				&aws.AwsRouteTable{
					Id: "table_from_state",
					Route: &[]struct {
						CidrBlock              *string `cty:"cidr_block"`
						EgressOnlyGatewayId    *string `cty:"egress_only_gateway_id"`
						GatewayId              *string `cty:"gateway_id"`
						InstanceId             *string `cty:"instance_id"`
						Ipv6CidrBlock          *string `cty:"ipv6_cidr_block"`
						LocalGatewayId         *string `cty:"local_gateway_id"`
						NatGatewayId           *string `cty:"nat_gateway_id"`
						NetworkInterfaceId     *string `cty:"network_interface_id"`
						TransitGatewayId       *string `cty:"transit_gateway_id"`
						VpcEndpointId          *string `cty:"vpc_endpoint_id"`
						VpcPeeringConnectionId *string `cty:"vpc_peering_connection_id"`
					}{
						{
							CidrBlock:     awssdk.String("0.0.0.0/0"),
							GatewayId:     awssdk.String("igw-07b7844a8fd17a638"),
							VpcEndpointId: awssdk.String(""),
						},
						{
							GatewayId:     awssdk.String("igw-07b7844a8fd17a638"),
							Ipv6CidrBlock: awssdk.String("::/0"),
						},
					},
				},
			},
			[]resource.Resource{
				&resource2.FakeResource{
					Id: "fake_resource",
				},
				&aws.AwsRouteTable{
					Id:    "table_from_state",
					Route: nil,
				},
				&aws.AwsRoute{
					Id:                      "r-table_from_state1080289494",
					RouteTableId:            awssdk.String("table_from_state"),
					DestinationCidrBlock:    awssdk.String("0.0.0.0/0"),
					GatewayId:               awssdk.String("igw-07b7844a8fd17a638"),
					Origin:                  awssdk.String("CreateRoute"),
					State:                   awssdk.String("active"),
					DestinationPrefixListId: awssdk.String(""),
					InstanceOwnerId:         awssdk.String(""),
				},
				&aws.AwsRoute{
					Id:                       "r-table_from_state2750132062",
					RouteTableId:             awssdk.String("table_from_state"),
					DestinationIpv6CidrBlock: awssdk.String("::/0"),
					GatewayId:                awssdk.String("igw-07b7844a8fd17a638"),
					Origin:                   awssdk.String("CreateRoute"),
					State:                    awssdk.String("active"),
					DestinationPrefixListId:  awssdk.String(""),
					InstanceOwnerId:          awssdk.String(""),
				},
			},
		},
		{
			"test route are expanded on default route tables",
			[]resource.Resource{
				&resource2.FakeResource{
					Id: "fake_resource",
				},
				&aws.AwsDefaultRouteTable{
					Id: "default_route_table_from_state",
					Route: &[]struct {
						CidrBlock              *string `cty:"cidr_block"`
						EgressOnlyGatewayId    *string `cty:"egress_only_gateway_id"`
						GatewayId              *string `cty:"gateway_id"`
						InstanceId             *string `cty:"instance_id"`
						Ipv6CidrBlock          *string `cty:"ipv6_cidr_block"`
						NatGatewayId           *string `cty:"nat_gateway_id"`
						NetworkInterfaceId     *string `cty:"network_interface_id"`
						TransitGatewayId       *string `cty:"transit_gateway_id"`
						VpcEndpointId          *string `cty:"vpc_endpoint_id"`
						VpcPeeringConnectionId *string `cty:"vpc_peering_connection_id"`
					}{
						{
							CidrBlock:     awssdk.String("0.0.0.0/0"),
							GatewayId:     awssdk.String("igw-07b7844a8fd17a638"),
							VpcEndpointId: awssdk.String(""),
						},
						{
							GatewayId:     awssdk.String("igw-07b7844a8fd17a638"),
							Ipv6CidrBlock: awssdk.String("::/0"),
						},
					},
				},
			},
			[]resource.Resource{
				&resource2.FakeResource{
					Id: "fake_resource",
				},
				&aws.AwsDefaultRouteTable{
					Id:    "default_route_table_from_state",
					Route: nil,
				},
				&aws.AwsRoute{
					Id:                      "r-default_route_table_from_state1080289494",
					RouteTableId:            awssdk.String("default_route_table_from_state"),
					DestinationCidrBlock:    awssdk.String("0.0.0.0/0"),
					GatewayId:               awssdk.String("igw-07b7844a8fd17a638"),
					Origin:                  awssdk.String("CreateRoute"),
					State:                   awssdk.String("active"),
					DestinationPrefixListId: awssdk.String(""),
					InstanceOwnerId:         awssdk.String(""),
				},
				&aws.AwsRoute{
					Id:                       "r-default_route_table_from_state2750132062",
					RouteTableId:             awssdk.String("default_route_table_from_state"),
					DestinationIpv6CidrBlock: awssdk.String("::/0"),
					GatewayId:                awssdk.String("igw-07b7844a8fd17a638"),
					Origin:                   awssdk.String("CreateRoute"),
					State:                    awssdk.String("active"),
					DestinationPrefixListId:  awssdk.String(""),
					InstanceOwnerId:          awssdk.String(""),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockedAlerter := &mocks.AlerterInterface{}
			m := NewAwsRouteTableExpander(mockedAlerter)
			err := m.Execute(nil, &tt.input)
			if err != nil {
				t.Fatal(err)
			}

			changelog, err := diff.Diff(tt.expected, tt.input)
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

func TestAwsRouteTableExpander_ExecuteWithInvalidRoutes(t *testing.T) {

	mockedAlerter := &mocks.AlerterInterface{}
	mockedAlerter.On("SendAlert", aws.AwsRouteTableResourceType, alerter.Alert{
		Message: "Skipped invalid route found in state for aws_route_table.table_from_state",
	})
	mockedAlerter.On("SendAlert", aws.AwsDefaultRouteTableResourceType, alerter.Alert{
		Message: "Skipped invalid route found in state for aws_default_route_table.default_table_from_state",
	})

	input := []resource.Resource{
		&aws.AwsRouteTable{
			Id: "table_from_state",
			Route: &[]struct {
				CidrBlock              *string `cty:"cidr_block"`
				EgressOnlyGatewayId    *string `cty:"egress_only_gateway_id"`
				GatewayId              *string `cty:"gateway_id"`
				InstanceId             *string `cty:"instance_id"`
				Ipv6CidrBlock          *string `cty:"ipv6_cidr_block"`
				LocalGatewayId         *string `cty:"local_gateway_id"`
				NatGatewayId           *string `cty:"nat_gateway_id"`
				NetworkInterfaceId     *string `cty:"network_interface_id"`
				TransitGatewayId       *string `cty:"transit_gateway_id"`
				VpcEndpointId          *string `cty:"vpc_endpoint_id"`
				VpcPeeringConnectionId *string `cty:"vpc_peering_connection_id"`
			}{
				{
					GatewayId: awssdk.String("igw-07b7844a8fd17a638"),
				},
			},
		},
		&aws.AwsDefaultRouteTable{
			Id: "default_table_from_state",
			Route: &[]struct {
				CidrBlock              *string `cty:"cidr_block"`
				EgressOnlyGatewayId    *string `cty:"egress_only_gateway_id"`
				GatewayId              *string `cty:"gateway_id"`
				InstanceId             *string `cty:"instance_id"`
				Ipv6CidrBlock          *string `cty:"ipv6_cidr_block"`
				NatGatewayId           *string `cty:"nat_gateway_id"`
				NetworkInterfaceId     *string `cty:"network_interface_id"`
				TransitGatewayId       *string `cty:"transit_gateway_id"`
				VpcEndpointId          *string `cty:"vpc_endpoint_id"`
				VpcPeeringConnectionId *string `cty:"vpc_peering_connection_id"`
			}{
				{
					GatewayId: awssdk.String("igw-07b7844a8fd17a638"),
				},
			},
		},
	}

	expected := []resource.Resource{
		&aws.AwsRouteTable{
			Id:    "table_from_state",
			Route: nil,
		},
		&aws.AwsDefaultRouteTable{
			Id:    "default_table_from_state",
			Route: nil,
		},
	}

	m := NewAwsRouteTableExpander(mockedAlerter)
	err := m.Execute(nil, &input)
	if err != nil {
		t.Fatal(err)
	}

	changelog, err := diff.Diff(expected, input)
	if err != nil {
		t.Fatal(err)
	}
	if len(changelog) > 0 {
		for _, change := range changelog {
			t.Errorf("%s got = %v, want %v", strings.Join(change.Path, "."), awsutil.Prettify(change.From), awsutil.Prettify(change.To))
		}
	}

	mockedAlerter.AssertExpectations(t)
}
