package middlewares

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/r3labs/diff/v2"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

func TestAwsDefaultNetworkACL_Execute(t *testing.T) {
	tests := []struct {
		name               string
		remoteResources    []*resource.Resource
		resourcesFromState []*resource.Resource
		expected           []*resource.Resource
	}{
		{
			"default network ACL is not ignored when managed by IaC",
			[]*resource.Resource{
				{
					Id: "fake",
				},
				{
					Id:   "default-acl",
					Type: aws.AwsDefaultNetworkACLResourceType,
				},
				{
					Id:   "non-default-acl",
					Type: aws.AwsNetworkACLResourceType,
				},
			},
			[]*resource.Resource{
				{
					Id:   "default-acl",
					Type: aws.AwsDefaultNetworkACLResourceType,
				},
			},
			[]*resource.Resource{
				{
					Id: "fake",
				},
				{
					Id:   "default-acl",
					Type: aws.AwsDefaultNetworkACLResourceType,
				},
				{
					Id:   "non-default-acl",
					Type: aws.AwsNetworkACLResourceType,
				},
			},
		},
		{
			"default network acl is ignored when not managed by IaC",
			[]*resource.Resource{
				{
					Id: "fake",
				},
				{
					Id:   "default-acl",
					Type: aws.AwsDefaultNetworkACLResourceType,
				},
				{
					Id:   "non-default-acl",
					Type: aws.AwsNetworkACLResourceType,
				},
			},
			[]*resource.Resource{},
			[]*resource.Resource{
				{
					Id: "fake",
				},
				{
					Id:   "non-default-acl",
					Type: aws.AwsNetworkACLResourceType,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewAwsDefaultNetworkACL()
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
