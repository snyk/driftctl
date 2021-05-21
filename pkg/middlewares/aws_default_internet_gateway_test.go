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

func TestAwsDefaultInternetGateway_Execute(t *testing.T) {
	tests := []struct {
		name               string
		remoteResources    []resource.Resource
		resourcesFromState []resource.Resource
		expected           []resource.Resource
	}{
		{
			"default internet gateway is not ignored when managed by IaC",
			[]resource.Resource{
				&resource2.FakeResource{
					Id: "fake",
				},
				&resource.AbstractResource{
					Id:   "default-vpc",
					Type: aws.AwsDefaultVpcResourceType,
				},
				&resource.AbstractResource{
					Id:   "dummy-vpc",
					Type: aws.AwsVpcResourceType,
				},
				&resource.AbstractResource{
					Id:   "default-igw",
					Type: aws.AwsInternetGatewayResourceType,
					Attrs: &resource.Attributes{
						"vpc_id": "default-vpc",
					},
				},
				&resource.AbstractResource{
					Id:   "dummy-igw",
					Type: aws.AwsInternetGatewayResourceType,
					Attrs: &resource.Attributes{
						"vpc_id": "dummy-vpc",
					},
				},
			},
			[]resource.Resource{
				&resource.AbstractResource{
					Id:   "default-igw",
					Type: aws.AwsInternetGatewayResourceType,
					Attrs: &resource.Attributes{
						"vpc_id": "default-vpc",
					},
				},
			},
			[]resource.Resource{
				&resource2.FakeResource{
					Id: "fake",
				},
				&resource.AbstractResource{
					Id:   "default-vpc",
					Type: aws.AwsDefaultVpcResourceType,
				},
				&resource.AbstractResource{
					Id:   "dummy-vpc",
					Type: aws.AwsVpcResourceType,
				},
				&resource.AbstractResource{
					Id:   "default-igw",
					Type: aws.AwsInternetGatewayResourceType,
					Attrs: &resource.Attributes{
						"vpc_id": "default-vpc",
					},
				},
				&resource.AbstractResource{
					Id:   "dummy-igw",
					Type: aws.AwsInternetGatewayResourceType,
					Attrs: &resource.Attributes{
						"vpc_id": "dummy-vpc",
					},
				},
			},
		},
		{
			"default internet gateway is ignored when not managed by IaC",
			[]resource.Resource{
				&resource2.FakeResource{
					Id: "fake",
				},
				&resource.AbstractResource{
					Id:   "default-vpc",
					Type: aws.AwsDefaultVpcResourceType,
				},
				&resource.AbstractResource{
					Id:   "dummy-vpc",
					Type: aws.AwsVpcResourceType,
				},
				&resource.AbstractResource{
					Id:   "default-igw",
					Type: aws.AwsInternetGatewayResourceType,
					Attrs: &resource.Attributes{
						"vpc_id": "default-vpc",
					},
				},
				&resource.AbstractResource{
					Id:   "dummy-igw",
					Type: aws.AwsInternetGatewayResourceType,
					Attrs: &resource.Attributes{
						"vpc_id": "dummy-vpc",
					},
				},
			},
			[]resource.Resource{},
			[]resource.Resource{
				&resource2.FakeResource{
					Id: "fake",
				},
				&resource.AbstractResource{
					Id:   "default-vpc",
					Type: aws.AwsDefaultVpcResourceType,
				},
				&resource.AbstractResource{
					Id:   "dummy-vpc",
					Type: aws.AwsVpcResourceType,
				},
				&resource.AbstractResource{
					Id:   "dummy-igw",
					Type: aws.AwsInternetGatewayResourceType,
					Attrs: &resource.Attributes{
						"vpc_id": "dummy-vpc",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewAwsDefaultInternetGateway()
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
