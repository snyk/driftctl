package middlewares

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/r3labs/diff/v2"
	"github.com/stretchr/testify/mock"

	"github.com/cloudskiff/driftctl/mocks"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	resource2 "github.com/cloudskiff/driftctl/test/resource"
)

func TestAwsRouteTableExpander_Execute(t *testing.T) {
	tests := []struct {
		name     string
		input    []resource.Resource
		expected []resource.Resource
		mock     func(factory *terraform.MockResourceFactory)
	}{
		{
			name: "test with nil route attributes",
			input: []resource.Resource{
				&resource.AbstractResource{
					Id:   "table_from_state",
					Type: aws.AwsRouteTableResourceType,
					Attrs: &resource.Attributes{
						"route": nil,
					},
				},
			},
			expected: []resource.Resource{
				&resource.AbstractResource{
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
			input: []resource.Resource{
				&resource.AbstractResource{
					Id:   "table_from_state",
					Type: aws.AwsRouteTableResourceType,
					Attrs: &resource.Attributes{
						"route": []interface{}{},
					},
				},
			},
			expected: []resource.Resource{
				&resource.AbstractResource{
					Id:    "table_from_state",
					Type:  aws.AwsRouteTableResourceType,
					Attrs: &resource.Attributes{},
				},
			},
		},
		{
			name: "test route are expanded",
			input: []resource.Resource{
				&resource2.FakeResource{
					Id: "fake_resource",
				},
				&resource.AbstractResource{
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
			expected: []resource.Resource{
				&resource2.FakeResource{
					Id: "fake_resource",
				},
				&resource.AbstractResource{
					Id:    "table_from_state",
					Type:  aws.AwsRouteTableResourceType,
					Attrs: &resource.Attributes{},
				},
				&resource.AbstractResource{
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
				&resource.AbstractResource{
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
			mock: func(factory *terraform.MockResourceFactory) {
				factory.On("CreateAbstractResource", "aws_route", mock.Anything, mock.MatchedBy(func(input map[string]interface{}) bool {
					return input["id"] == "r-table_from_state1080289494"
				})).Times(1).Return(&resource.AbstractResource{
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
				})).Times(1).Return(&resource.AbstractResource{
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
		{
			name: "test route are expanded on default route tables",
			input: []resource.Resource{
				&resource2.FakeResource{
					Id: "fake_resource",
				},
				&resource.AbstractResource{
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
			expected: []resource.Resource{
				&resource2.FakeResource{
					Id: "fake_resource",
				},
				&resource.AbstractResource{
					Id:    "default_route_table_from_state",
					Type:  aws.AwsDefaultRouteTableResourceType,
					Attrs: &resource.Attributes{},
				},
				&resource.AbstractResource{
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
				&resource.AbstractResource{
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
			mock: func(factory *terraform.MockResourceFactory) {
				factory.On("CreateAbstractResource", "aws_route", mock.Anything, mock.MatchedBy(func(input map[string]interface{}) bool {
					return input["id"] == "r-default_route_table_from_state1080289494"
				})).Times(1).Return(&resource.AbstractResource{
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
				})).Times(1).Return(&resource.AbstractResource{
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
			[]resource.Resource{
				&resource2.FakeResource{
					Id: "fake_resource",
				},
				&resource.AbstractResource{
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
				&resource.AbstractResource{
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
			[]resource.Resource{
				&resource2.FakeResource{
					Id: "fake_resource",
				},
				&resource.AbstractResource{
					Id:    "default_route_table_from_state",
					Type:  aws.AwsDefaultRouteTableResourceType,
					Attrs: &resource.Attributes{},
				},
				&resource.AbstractResource{
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
				&resource.AbstractResource{
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
			func(factory *terraform.MockResourceFactory) {
				factory.On("CreateAbstractResource", "aws_route", mock.Anything, mock.MatchedBy(func(input map[string]interface{}) bool {
					return input["id"] == "r-default_route_table_from_state1080289494"
				})).Times(1).Return(&resource.AbstractResource{
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
			[]resource.Resource{
				&resource2.FakeResource{
					Id: "fake_resource",
				},
				&resource.AbstractResource{
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
				&resource.AbstractResource{
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
			[]resource.Resource{
				&resource2.FakeResource{
					Id: "fake_resource",
				},
				&resource.AbstractResource{
					Id:    "table_from_state",
					Type:  aws.AwsRouteTableResourceType,
					Attrs: &resource.Attributes{},
				},
				&resource.AbstractResource{
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
				&resource.AbstractResource{
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
			func(factory *terraform.MockResourceFactory) {
				factory.On("CreateAbstractResource", "aws_route", mock.Anything, mock.MatchedBy(func(input map[string]interface{}) bool {
					return input["id"] == "r-table_from_state2750132062"
				})).Times(1).Return(&resource.AbstractResource{
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

			factory := &terraform.MockResourceFactory{}
			if tt.mock != nil {
				tt.mock(factory)
			}

			m := NewAwsRouteTableExpander(mockedAlerter, factory)
			err := m.Execute(&[]resource.Resource{}, &tt.input)
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
	mockedAlerter.On("SendAlert", aws.AwsRouteTableResourceType, newInvalidRouteAlert(
		"aws_route_table", "table_from_state",
	))
	mockedAlerter.On("SendAlert", aws.AwsDefaultRouteTableResourceType, newInvalidRouteAlert(
		"aws_default_route_table", "default_table_from_state",
	))

	input := []resource.Resource{
		&resource.AbstractResource{
			Id:   "table_from_state",
			Type: aws.AwsRouteTableResourceType,
			Attrs: &resource.Attributes{
				"route": []interface{}{
					map[string]interface{}{
						"gateway_id":      "igw-07b7844a8fd17a638",
						"cidr_block":      "",
						"ipv6_cidr_block": "",
					},
				},
			},
		},
		&resource.AbstractResource{
			Id:   "default_table_from_state",
			Type: aws.AwsDefaultRouteTableResourceType,
			Attrs: &resource.Attributes{
				"route": []interface{}{
					map[string]interface{}{
						"gateway_id":      "igw-07b7844a8fd17a638",
						"cidr_block":      "",
						"ipv6_cidr_block": "",
					},
				},
			},
		},
	}

	expected := []resource.Resource{
		&resource.AbstractResource{
			Id:    "table_from_state",
			Type:  aws.AwsRouteTableResourceType,
			Attrs: &resource.Attributes{},
		},
		&resource.AbstractResource{
			Id:    "default_table_from_state",
			Type:  aws.AwsDefaultRouteTableResourceType,
			Attrs: &resource.Attributes{},
		},
	}

	factory := &terraform.MockResourceFactory{}

	m := NewAwsRouteTableExpander(mockedAlerter, factory)
	err := m.Execute(&[]resource.Resource{}, &input)
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
