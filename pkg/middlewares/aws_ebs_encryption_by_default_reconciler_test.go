package middlewares

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/r3labs/diff/v2"
	"github.com/snyk/driftctl/pkg/terraform"

	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

func TestAwsEbsEncryptionByDefaultReconciler_Execute(t *testing.T) {
	tests := []struct {
		name               string
		mocks              func(*terraform.MockResourceFactory)
		remoteResources    []*resource.Resource
		resourcesFromState []*resource.Resource
		expected           []*resource.Resource
	}{
		{
			name: "test encryption by default is managed",
			mocks: func(factory *terraform.MockResourceFactory) {
				factory.On("CreateAbstractResource",
					aws.AwsEbsEncryptionByDefaultResourceType,
					"terraform-20220328091515068500000001",
					map[string]interface{}{
						"id":      "terraform-20220328091515068500000001",
						"enabled": true,
					}).Return(&resource.Resource{
					Id:   "terraform-20220328091515068500000001",
					Type: aws.AwsEbsEncryptionByDefaultResourceType,
					Attrs: &resource.Attributes{
						"id":      "terraform-20220328091515068500000001",
						"enabled": true,
					},
				}).Once()
			},
			remoteResources: []*resource.Resource{
				{
					Id:    "bucket-1",
					Type:  aws.AwsS3BucketResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:   "test-encryption",
					Type: aws.AwsEbsEncryptionByDefaultResourceType,
					Attrs: &resource.Attributes{
						"enabled": true,
					},
				},
			},
			resourcesFromState: []*resource.Resource{
				{
					Id:    "bucket-1",
					Type:  aws.AwsS3BucketResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:   "terraform-20220328091515068500000001",
					Type: aws.AwsEbsEncryptionByDefaultResourceType,
					Attrs: &resource.Attributes{
						"id":      "terraform-20220328091515068500000001",
						"enabled": true,
					},
				},
			},
			expected: []*resource.Resource{
				{
					Id:    "bucket-1",
					Type:  aws.AwsS3BucketResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:   "terraform-20220328091515068500000001",
					Type: aws.AwsEbsEncryptionByDefaultResourceType,
					Attrs: &resource.Attributes{
						"id":      "terraform-20220328091515068500000001",
						"enabled": true,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := &terraform.MockResourceFactory{}
			if tt.mocks != nil {
				tt.mocks(factory)
			}

			m := NewAwsEbsEncryptionByDefaultReconciler(factory)
			err := m.Execute(&tt.remoteResources, &tt.resourcesFromState)
			if err != nil {
				t.Fatal(err)
			}

			changelog, err := diff.Diff(tt.remoteResources, tt.expected)
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
