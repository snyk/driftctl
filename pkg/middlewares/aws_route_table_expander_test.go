package middlewares

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/r3labs/diff/v2"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/mocks"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
	"github.com/stretchr/testify/mock"
)

func TestAwsRouteTableExpander_Execute(t *testing.T) {
	tests := []struct {
		name     string
		input    []*resource.Resource
		expected []*resource.Resource
		mock     func(factory *dctlresource.MockResourceFactory)
	}{
		{
			name: "test with nil route attributes",
			input: []*resource.Resource{
				{
					Id:   "table_from_state",
					Type: aws.AwsRouteTableResourceType,
					Attrs: &resource.Attributes{
						"route": nil,
					},
				},
			},
			expected: []*resource.Resource{
				{
					Id:   "table_from_state",
					Type: aws.AwsRouteTableResourceType,
					Attrs: &resource.Attributes{
						"route": nil,
					},
				},
			},
		},
		{
			name: "test with empty route attributes",
			input: []*resource.Resource{
				{
					Id:   "table_from_state",
					Type: aws.AwsRouteTableResourceType,
					Attrs: &resource.Attributes{
						"route": []interface{}{},
					},
				},
			},
			expected: []*resource.Resource{
				{
					Id:    "table_from_state",
					Type:  aws.AwsRouteTableResourceType,
					Attrs: &resource.Attributes{},
				},
			},
		},
		{
			name: "test route are expanded",
			input: []*resource.Resource{
				{
					Id: "fake_resource",
				},
				{
					Id:   "table_from_state",
					Type: aws.AwsRouteTableResourceType,
					Attrs: &resource.Attributes{
						"route": []interface{}{
							map[string]interface{}{
								"gateway_id":      "igw-07b7844a8fd17a638",
								"cidr_block":      "0.0.0.0/0",
								"ipv6_cidr_block": "",
								"vpc_endpoint_id": "",
							},
							map[string]interface{}{
								"gateway_id":      "igw-07b7844a8fd17a638",
								"cidr_block":      "",
								"ipv6_cidr_block": "::/0",
							},
							map[string]interface{}{
								"gateway_id":                 "igw-07b7844a8fd17a638",
								"destination_prefix_list_id": "pl-63a5400a",
							},
						},
					},
				},
			},
			expected: []*resource.Resource{
				{
					Id: "fake_resource",
				},
				{
					Id:    "table_from_state",
					Type:  aws.AwsRouteTableResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:   "r-table_from_state1080289494",
					Type: aws.AwsRouteResourceType,
					Attrs: &resource.Attributes{
						"route_table_id":             "table_from_state",
						"origin":                     "CreateRoute",
						"destination_cidr_block":     "0.0.0.0/0",
						"gateway_id":                 "igw-07b7844a8fd17a638",
						"state":                      "active",
						"destination_prefix_list_id": "",
						"instance_owner_id":          "",
					},
				},
				&resource.Resource{
					Id:   "r-table_from_state2750132062",
					Type: aws.AwsRouteResourceType,
					Attrs: &resource.Attributes{
						"route_table_id":              "table_from_state",
						"origin":                      "CreateRoute",
						"destination_ipv6_cidr_block": "::/0",
						"gateway_id":                  "igw-07b7844a8fd17a638",
						"state":                       "active",
						"destination_prefix_list_id":  "",
						"instance_owner_id":           "",
					},
				},
				&resource.Resource{
					Id:   "r-table_from_state3813769586",
					Type: aws.AwsRouteResourceType,
					Attrs: &resource.Attributes{
						"route_table_id":             "table_from_state",
						"origin":                     "CreateRoute",
						"gateway_id":                 "igw-07b7844a8fd17a638",
						"state":                      "active",
						"destination_prefix_list_id": "pl-63a5400a",
						"instance_owner_id":          "",
					},
				},
			},
			mock: func(factory *dctlresource.MockResourceFactory) {
				factory.On("CreateAbstractResource", "aws_route", mock.Anything, mock.MatchedBy(func(input map[string]interface{}) bool {
					return input["id"] == "r-table_from_state1080289494"
				})).Times(1).Return(&resource.Resource{
					Id:   "r-table_from_state1080289494",
					Type: aws.AwsRouteResourceType,
					Attrs: &resource.Attributes{
						"route_table_id":             "table_from_state",
						"origin":                     "CreateRoute",
						"destination_cidr_block":     "0.0.0.0/0",
						"gateway_id":                 "igw-07b7844a8fd17a638",
						"state":                      "active",
						"destination_prefix_list_id": "",
						"instance_owner_id":          "",
					},
				}, nil)
				factory.On("CreateAbstractResource", "aws_route", mock.Anything, mock.MatchedBy(func(input map[string]interface{}) bool {
					return input["id"] == "r-table_from_state2750132062"
				})).Times(1).Return(&resource.Resource{
					Id:   "r-table_from_state2750132062",
					Type: aws.AwsRouteResourceType,
					Attrs: &resource.Attributes{
						"route_table_id":              "table_from_state",
						"origin":                      "CreateRoute",
						"destination_ipv6_cidr_block": "::/0",
						"gateway_id":                  "igw-07b7844a8fd17a638",
						"state":                       "active",
						"destination_prefix_list_id":  "",
						"instance_owner_id":           "",
					},
				}, nil)
				factory.On("CreateAbstractResource", "aws_route", mock.Anything, mock.MatchedBy(func(input map[string]interface{}) bool {
					return input["id"] == "r-table_from_state3813769586"
				})).Times(1).Return(&resource.Resource{
					Id:   "r-table_from_state3813769586",
					Type: aws.AwsRouteResourceType,
					Attrs: &resource.Attributes{
						"route_table_id":             "table_from_state",
						"origin":                     "CreateRoute",
						"gateway_id":                 "igw-07b7844a8fd17a638",
						"state":                      "active",
						"destination_prefix_list_id": "pl-63a5400a",
						"instance_owner_id":          "",
					},
				}, nil)
			},
		},
		{
			name: "test route are expanded on default route tables",
			input: []*resource.Resource{
				&resource.Resource{
					Id: "fake_resource",
				},
				&resource.Resource{
					Id:   "default_route_table_from_state",
					Type: aws.AwsDefaultRouteTableResourceType,
					Attrs: &resource.Attributes{
						"route": []interface{}{
							map[string]interface{}{
								"gateway_id":      "igw-07b7844a8fd17a638",
								"cidr_block":      "0.0.0.0/0",
								"ipv6_cidr_block": "",
								"vpc_endpoint_id": "",
							},
							map[string]interface{}{
								"gateway_id":      "igw-07b7844a8fd17a638",
								"cidr_block":      "",
								"ipv6_cidr_block": "::/0",
							},
						},
					},
				},
			},
			expected: []*resource.Resource{
				&resource.Resource{
					Id: "fake_resource",
				},
				&resource.Resource{
					Id:    "default_route_table_from_state",
					Type:  aws.AwsDefaultRouteTableResourceType,
					Attrs: &resource.Attributes{},
				},
				&resource.Resource{
					Id:   "r-default_route_table_from_state1080289494",
					Type: aws.AwsRouteResourceType,
					Attrs: &resource.Attributes{
						"route_table_id":             "default_route_table_from_state",
						"origin":                     "CreateRoute",
						"destination_cidr_block":     "0.0.0.0/0",
						"gateway_id":                 "igw-07b7844a8fd17a638",
						"state":                      "active",
						"destination_prefix_list_id": "",
						"instance_owner_id":          "",
					},
				},
				&resource.Resource{
					Id:   "r-default_route_table_from_state2750132062",
					Type: aws.AwsRouteResourceType,
					Attrs: &resource.Attributes{
						"route_table_id":              "default_route_table_from_state",
						"origin":                      "CreateRoute",
						"destination_ipv6_cidr_block": "::/0",
						"gateway_id":                  "igw-07b7844a8fd17a638",
						"state":                       "active",
						"destination_prefix_list_id":  "",
						"instance_owner_id":           "",
					},
				},
			},
			mock: func(factory *dctlresource.MockResourceFactory) {
				factory.On("CreateAbstractResource", "aws_route", mock.Anything, mock.MatchedBy(func(input map[string]interface{}) bool {
					return input["id"] == "r-default_route_table_from_state1080289494"
				})).Times(1).Return(&resource.Resource{
					Id:   "r-default_route_table_from_state1080289494",
					Type: aws.AwsRouteResourceType,
					Attrs: &resource.Attributes{
						"route_table_id":             "default_route_table_from_state",
						"origin":                     "CreateRoute",
						"destination_cidr_block":     "0.0.0.0/0",
						"gateway_id":                 "igw-07b7844a8fd17a638",
						"state":                      "active",
						"destination_prefix_list_id": "",
						"instance_owner_id":          "",
					},
				}, nil)
				factory.On("CreateAbstractResource", "aws_route", mock.Anything, mock.MatchedBy(func(input map[string]interface{}) bool {
					return input["id"] == "r-default_route_table_from_state2750132062"
				})).Times(1).Return(&resource.Resource{
					Id:   "r-default_route_table_from_state2750132062",
					Type: aws.AwsRouteResourceType,
					Attrs: &resource.Attributes{
						"route_table_id":              "default_route_table_from_state",
						"origin":                      "CreateRoute",
						"destination_ipv6_cidr_block": "::/0",
						"gateway_id":                  "igw-07b7844a8fd17a638",
						"state":                       "active",
						"destination_prefix_list_id":  "",
						"instance_owner_id":           "",
					},
				}, nil)
			},
		},
		{
			"test routes are expanded from default route tables except when they already exist",
			[]*resource.Resource{
				&resource.Resource{
					Id: "fake_resource",
				},
				&resource.Resource{
					Id:   "r-default_route_table_from_state2750132062",
					Type: aws.AwsRouteResourceType,
					Attrs: &resource.Attributes{
						"route_table_id":              "default_route_table_from_state",
						"origin":                      "CreateRoute",
						"destination_ipv6_cidr_block": "::/0",
						"gateway_id":                  "igw-07b7844a8fd17a638",
						"state":                       "active",
						"destination_prefix_list_id":  "",
						"instance_owner_id":           "",
					},
				},
				&resource.Resource{
					Id:   "default_route_table_from_state",
					Type: aws.AwsDefaultRouteTableResourceType,
					Attrs: &resource.Attributes{
						"route": []interface{}{
							map[string]interface{}{
								"gateway_id":      "igw-07b7844a8fd17a638",
								"cidr_block":      "0.0.0.0/0",
								"ipv6_cidr_block": "",
								"vpc_endpoint_id": "",
							},
							map[string]interface{}{
								"gateway_id":      "igw-07b7844a8fd17a638",
								"cidr_block":      "",
								"ipv6_cidr_block": "::/0",
							},
						},
					},
				},
			},
			[]*resource.Resource{
				&resource.Resource{
					Id: "fake_resource",
				},
				&resource.Resource{
					Id:    "default_route_table_from_state",
					Type:  aws.AwsDefaultRouteTableResourceType,
					Attrs: &resource.Attributes{},
				},
				&resource.Resource{
					Id:   "r-default_route_table_from_state1080289494",
					Type: aws.AwsRouteResourceType,
					Attrs: &resource.Attributes{
						"route_table_id":             "default_route_table_from_state",
						"origin":                     "CreateRoute",
						"destination_cidr_block":     "0.0.0.0/0",
						"gateway_id":                 "igw-07b7844a8fd17a638",
						"state":                      "active",
						"destination_prefix_list_id": "",
						"instance_owner_id":          "",
					},
				},
				&resource.Resource{
					Id:   "r-default_route_table_from_state2750132062",
					Type: aws.AwsRouteResourceType,
					Attrs: &resource.Attributes{
						"route_table_id":              "default_route_table_from_state",
						"origin":                      "CreateRoute",
						"destination_ipv6_cidr_block": "::/0",
						"gateway_id":                  "igw-07b7844a8fd17a638",
						"state":                       "active",
						"destination_prefix_list_id":  "",
						"instance_owner_id":           "",
					},
				},
			},
			func(factory *dctlresource.MockResourceFactory) {
				factory.On("CreateAbstractResource", "aws_route", mock.Anything, mock.MatchedBy(func(input map[string]interface{}) bool {
					return input["id"] == "r-default_route_table_from_state1080289494"
				})).Times(1).Return(&resource.Resource{
					Id:   "r-default_route_table_from_state1080289494",
					Type: aws.AwsRouteResourceType,
					Attrs: &resource.Attributes{
						"route_table_id":             "default_route_table_from_state",
						"origin":                     "CreateRoute",
						"destination_cidr_block":     "0.0.0.0/0",
						"gateway_id":                 "igw-07b7844a8fd17a638",
						"state":                      "active",
						"destination_prefix_list_id": "",
						"instance_owner_id":          "",
					},
				}, nil)
			},
		},
		{
			"test routes are expanded except when they already exist",
			[]*resource.Resource{
				&resource.Resource{
					Id: "fake_resource",
				},
				&resource.Resource{
					Id:   "r-table_from_state1080289494",
					Type: aws.AwsRouteResourceType,
					Attrs: &resource.Attributes{
						"route_table_id":             "table_from_state",
						"origin":                     "CreateRoute",
						"destination_cidr_block":     "0.0.0.0/0",
						"gateway_id":                 "igw-07b7844a8fd17a638",
						"state":                      "active",
						"destination_prefix_list_id": "",
						"instance_owner_id":          "",
					},
				},
				&resource.Resource{
					Id:   "table_from_state",
					Type: aws.AwsRouteTableResourceType,
					Attrs: &resource.Attributes{
						"route": []interface{}{
							map[string]interface{}{
								"gateway_id":      "igw-07b7844a8fd17a638",
								"cidr_block":      "0.0.0.0/0",
								"ipv6_cidr_block": "",
								"vpc_endpoint_id": "",
							},
							map[string]interface{}{
								"gateway_id":      "igw-07b7844a8fd17a638",
								"cidr_block":      "",
								"ipv6_cidr_block": "::/0",
							},
						},
					},
				},
			},
			[]*resource.Resource{
				&resource.Resource{
					Id: "fake_resource",
				},
				&resource.Resource{
					Id:    "table_from_state",
					Type:  aws.AwsRouteTableResourceType,
					Attrs: &resource.Attributes{},
				},
				&resource.Resource{
					Id:   "r-table_from_state1080289494",
					Type: aws.AwsRouteResourceType,
					Attrs: &resource.Attributes{
						"route_table_id":             "table_from_state",
						"origin":                     "CreateRoute",
						"destination_cidr_block":     "0.0.0.0/0",
						"gateway_id":                 "igw-07b7844a8fd17a638",
						"state":                      "active",
						"destination_prefix_list_id": "",
						"instance_owner_id":          "",
					},
				},
				&resource.Resource{
					Id:   "r-table_from_state2750132062",
					Type: aws.AwsRouteResourceType,
					Attrs: &resource.Attributes{
						"route_table_id":              "table_from_state",
						"origin":                      "CreateRoute",
						"destination_ipv6_cidr_block": "::/0",
						"gateway_id":                  "igw-07b7844a8fd17a638",
						"state":                       "active",
						"destination_prefix_list_id":  "",
						"instance_owner_id":           "",
					},
				},
			},
			func(factory *dctlresource.MockResourceFactory) {
				factory.On("CreateAbstractResource", "aws_route", mock.Anything, mock.MatchedBy(func(input map[string]interface{}) bool {
					return input["id"] == "r-table_from_state2750132062"
				})).Times(1).Return(&resource.Resource{
					Id:   "r-table_from_state2750132062",
					Type: aws.AwsRouteResourceType,
					Attrs: &resource.Attributes{
						"route_table_id":              "table_from_state",
						"origin":                      "CreateRoute",
						"destination_ipv6_cidr_block": "::/0",
						"gateway_id":                  "igw-07b7844a8fd17a638",
						"state":                       "active",
						"destination_prefix_list_id":  "",
						"instance_owner_id":           "",
					},
				}, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockedAlerter := &mocks.AlerterInterface{}

			factory := &dctlresource.MockResourceFactory{}
			if tt.mock != nil {
				tt.mock(factory)
			}

			m := NewAwsRouteTableExpander(mockedAlerter, factory)
			err := m.Execute(&[]*resource.Resource{}, &tt.input)
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
