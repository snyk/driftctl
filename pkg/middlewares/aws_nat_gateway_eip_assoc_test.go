package middlewares

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/r3labs/diff/v2"
)

func TestAwsNatGatewayEipAssoc_Execute(t *testing.T) {
	tests := []struct {
		name     string
		input    []*resource.Resource
		expected []*resource.Resource
	}{
		{
			name: "test nil values do not crash middleware",
			input: []*resource.Resource{
				{
					Id:    "nat-0a5408508b19ef490",
					Type:  aws.AwsNatGatewayResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Type:  aws.AwsEipAssociationResourceType,
					Id:    "eipassoc-0d32af6acf31df913",
					Attrs: &resource.Attributes{},
				},
			},
			expected: []*resource.Resource{
				{
					Id:    "nat-0a5408508b19ef490",
					Type:  aws.AwsNatGatewayResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Type:  aws.AwsEipAssociationResourceType,
					Id:    "eipassoc-0d32af6acf31df913",
					Attrs: &resource.Attributes{},
				},
			},
		},
		{
			name: "test eip assoc ignored when associated to a nat gateway",
			input: []*resource.Resource{
				{
					Type: aws.AwsNatGatewayResourceType,
					Attrs: &resource.Attributes{
						"allocation_id": "eipalloc-0f3e9fff457bb770b",
					},
				},
				{
					Type: aws.AwsEipAssociationResourceType,
					Attrs: &resource.Attributes{
						"allocation_id": "eipalloc-0f3e9fff457bb770b",
					},
				},
				{
					Type: aws.AwsEipAssociationResourceType,
					Attrs: &resource.Attributes{
						"allocation_id": "eipalloc-1234567890",
					},
				},
			},
			expected: []*resource.Resource{
				&resource.Resource{
					Type: aws.AwsNatGatewayResourceType,
					Attrs: &resource.Attributes{
						"allocation_id": "eipalloc-0f3e9fff457bb770b",
					},
				},
				&resource.Resource{
					Type: aws.AwsEipAssociationResourceType,
					Attrs: &resource.Attributes{
						"allocation_id": "eipalloc-1234567890",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewAwsNatGatewayEipAssoc()
			err := middleware.Execute(&tt.input, &[]*resource.Resource{})
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
