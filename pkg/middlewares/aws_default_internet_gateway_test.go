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
				&aws.AwsDefaultVpc{
					Id: "default-vpc",
				},
				&aws.AwsVpc{
					Id: "dummy-vpc",
				},
				&aws.AwsInternetGateway{
					Id:    "default-igw",
					VpcId: awssdk.String("default-vpc"),
				},
				&aws.AwsInternetGateway{
					Id:    "dummy-igw",
					VpcId: awssdk.String("dummy-vpc"),
				},
			},
			[]resource.Resource{
				&aws.AwsInternetGateway{
					Id:    "default-igw",
					VpcId: awssdk.String("default-vpc"),
				},
			},
			[]resource.Resource{
				&resource2.FakeResource{
					Id: "fake",
				},
				&aws.AwsDefaultVpc{
					Id: "default-vpc",
				},
				&aws.AwsVpc{
					Id: "dummy-vpc",
				},
				&aws.AwsInternetGateway{
					Id:    "default-igw",
					VpcId: awssdk.String("default-vpc"),
				},
				&aws.AwsInternetGateway{
					Id:    "dummy-igw",
					VpcId: awssdk.String("dummy-vpc"),
				},
			},
		},
		{
			"default internet gateway is ignored when not managed by IaC",
			[]resource.Resource{
				&resource2.FakeResource{
					Id: "fake",
				},
				&aws.AwsDefaultVpc{
					Id: "default-vpc",
				},
				&aws.AwsVpc{
					Id: "dummy-vpc",
				},
				&aws.AwsInternetGateway{
					Id:    "default-igw",
					VpcId: awssdk.String("default-vpc"),
				},
				&aws.AwsInternetGateway{
					Id:    "dummy-igw",
					VpcId: awssdk.String("dummy-vpc"),
				},
			},
			[]resource.Resource{},
			[]resource.Resource{
				&resource2.FakeResource{
					Id: "fake",
				},
				&aws.AwsDefaultVpc{
					Id: "default-vpc",
				},
				&aws.AwsVpc{
					Id: "dummy-vpc",
				},
				&aws.AwsInternetGateway{
					Id:    "dummy-igw",
					VpcId: awssdk.String("dummy-vpc"),
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
