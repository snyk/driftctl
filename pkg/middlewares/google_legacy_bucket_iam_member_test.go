package middlewares

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/google"
	"github.com/r3labs/diff/v2"
)

func TestGoogleLegacyBucketIAMBindings_Execute(t *testing.T) {
	tests := []struct {
		name               string
		remoteResources    []*resource.Resource
		resourcesFromState []*resource.Resource
		expected           []*resource.Resource
	}{
		{
			"test that non legacy bindings are not ignored when managed by IaC",
			[]*resource.Resource{
				{
					Id:    "fake",
					Type:  google.GoogleStorageBucketResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:   "admin bucket",
					Type: google.GoogleStorageBucketIamMemberResourceType,
					Attrs: &resource.Attributes{
						"role": "storage.admin",
					},
				},
				{
					Id:   "legacy",
					Type: google.GoogleStorageBucketIamMemberResourceType,
					Attrs: &resource.Attributes{
						"role": "storage.legacyBucketOwner",
					},
				},
			},
			[]*resource.Resource{},
			[]*resource.Resource{
				{
					Id:    "fake",
					Type:  google.GoogleStorageBucketResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:   "admin bucket",
					Type: google.GoogleStorageBucketIamMemberResourceType,
					Attrs: &resource.Attributes{
						"role": "storage.admin",
					},
				},
			},
		},
		{
			"test that legacy are not ignored when managed",
			[]*resource.Resource{
				{
					Id:    "fake",
					Type:  google.GoogleStorageBucketResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:   "admin bucket",
					Type: google.GoogleStorageBucketIamMemberResourceType,
					Attrs: &resource.Attributes{
						"role": "storage.admin",
					},
				},
				{
					Id:   "legacy",
					Type: google.GoogleStorageBucketIamMemberResourceType,
					Attrs: &resource.Attributes{
						"role": "storage.legacyBucketOwner",
					},
				},
				{
					Id:   "legacy-managed",
					Type: google.GoogleStorageBucketIamMemberResourceType,
					Attrs: &resource.Attributes{
						"role": "storage.legacyBucketOwner",
					},
				},
			},
			[]*resource.Resource{
				{
					Id:   "legacy-managed",
					Type: google.GoogleStorageBucketIamMemberResourceType,
					Attrs: &resource.Attributes{
						"role": "storage.legacyBucketOwner",
					},
				},
			},
			[]*resource.Resource{
				{
					Id:    "fake",
					Type:  google.GoogleStorageBucketResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:   "admin bucket",
					Type: google.GoogleStorageBucketIamMemberResourceType,
					Attrs: &resource.Attributes{
						"role": "storage.admin",
					},
				},
				{
					Id:   "legacy-managed",
					Type: google.GoogleStorageBucketIamMemberResourceType,
					Attrs: &resource.Attributes{
						"role": "storage.legacyBucketOwner",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewGoogleLegacyBucketIAMBindings()
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
