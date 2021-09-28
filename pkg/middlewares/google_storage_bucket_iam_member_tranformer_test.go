package middlewares

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/google"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/r3labs/diff/v2"
)

func TestGoogleBucketIAMMemberTransformer_Execute(t *testing.T) {
	tests := []struct {
		name               string
		resourcesFromState []*resource.Resource
		expected           []*resource.Resource
		mock               func(factory *terraform.MockResourceFactory)
	}{
		{
			"Test that bucket member are transformed into bindings",
			[]*resource.Resource{
				{
					Id:    "fake",
					Type:  google.GoogleStorageBucketResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:   "admin bucket",
					Type: google.GoogleStorageBucketIamBindingResourceType,
					Attrs: &resource.Attributes{
						"role": "storage.admin",
					},
				},
				{
					Id:   "b/bucket/admin/elie",
					Type: google.GoogleStorageBucketIamMemberResourceType,
					Attrs: &resource.Attributes{
						"role":   "storage.admin",
						"bucket": "b/bucket",
						"member": "user:elie@cloudskiff.com",
					},
				},
				{
					Id:   "b/bucket/admin/William",
					Type: google.GoogleStorageBucketIamMemberResourceType,
					Attrs: &resource.Attributes{
						"role":   "storage.admin",
						"bucket": "b/bucket",
						"member": "user:william@cloudskiff.com",
					},
				},
				{
					Id:   "b/bucket/viewer/William",
					Type: google.GoogleStorageBucketIamMemberResourceType,
					Attrs: &resource.Attributes{
						"role":   "storage.viewer",
						"bucket": "b/bucket",
						"member": "user:william@cloudskiff.com",
					},
				},
				{
					Id:   "b/bucket2/viewer/William",
					Type: google.GoogleStorageBucketIamMemberResourceType,
					Attrs: &resource.Attributes{
						"role":   "storage.viewer",
						"bucket": "b/bucket2",
						"member": "user:william@cloudskiff.com",
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
					Type: google.GoogleStorageBucketIamBindingResourceType,
					Attrs: &resource.Attributes{
						"role": "storage.admin",
					},
				},
				{
					Id:   "b/bucket/storage.admin",
					Type: google.GoogleStorageBucketIamBindingResourceType,
					Attrs: &resource.Attributes{
						"role":   "storage.admin",
						"bucket": "b/bucket",
						"members": []string{
							"user:elie@cloudskiff.com",
							"user:william@cloudskiff.com",
						},
					},
				},
				{
					Id:   "b/bucket/storage.viewer",
					Type: google.GoogleStorageBucketIamBindingResourceType,
					Attrs: &resource.Attributes{
						"role":   "storage.viewer",
						"bucket": "b/bucket",
						"members": []string{
							"user:william@cloudskiff.com",
						},
					},
				},
				{
					Id:   "b/bucket2/storage.viewer",
					Type: google.GoogleStorageBucketIamBindingResourceType,
					Attrs: &resource.Attributes{
						"role":   "storage.viewer",
						"bucket": "b/bucket2",
						"members": []string{
							"user:william@cloudskiff.com",
						},
					},
				},
			},
			func(factory *terraform.MockResourceFactory) {
				factory.On(
					"CreateAbstractResource", google.GoogleStorageBucketIamBindingResourceType,
					"b/bucket/storage.admin",
					map[string]interface{}{
						"id":     "b/bucket/storage.admin",
						"bucket": "b/bucket",
						"role":   "storage.admin",
						"members": []string{
							"user:elie@cloudskiff.com",
							"user:william@cloudskiff.com",
						},
					}).Return(&resource.Resource{
					Id:   "b/bucket/storage.admin",
					Type: google.GoogleStorageBucketIamBindingResourceType,
					Attrs: &resource.Attributes{
						"role":   "storage.admin",
						"bucket": "b/bucket",
						"members": []string{
							"user:elie@cloudskiff.com",
							"user:william@cloudskiff.com",
						},
					},
				}).Once()

				factory.On(
					"CreateAbstractResource", google.GoogleStorageBucketIamBindingResourceType,
					"b/bucket/storage.viewer",
					map[string]interface{}{
						"id":     "b/bucket/storage.viewer",
						"bucket": "b/bucket",
						"role":   "storage.viewer",
						"members": []string{
							"user:william@cloudskiff.com",
						},
					}).Return(&resource.Resource{
					Id:   "b/bucket/storage.viewer",
					Type: google.GoogleStorageBucketIamBindingResourceType,
					Attrs: &resource.Attributes{
						"role":   "storage.viewer",
						"bucket": "b/bucket",
						"members": []string{
							"user:william@cloudskiff.com",
						},
					},
				}).Once()

				factory.On(
					"CreateAbstractResource", google.GoogleStorageBucketIamBindingResourceType,
					"b/bucket2/storage.viewer",
					map[string]interface{}{
						"id":     "b/bucket2/storage.viewer",
						"bucket": "b/bucket2",
						"role":   "storage.viewer",
						"members": []string{
							"user:william@cloudskiff.com",
						},
					}).Return(&resource.Resource{
					Id:   "b/bucket2/storage.viewer",
					Type: google.GoogleStorageBucketIamBindingResourceType,
					Attrs: &resource.Attributes{
						"role":   "storage.viewer",
						"bucket": "b/bucket2",
						"members": []string{
							"user:william@cloudskiff.com",
						},
					},
				}).Once()
			},
		},
		{
			"test that everything is fine when there is no members",
			[]*resource.Resource{
				{
					Id:    "fake",
					Type:  google.GoogleStorageBucketResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:   "admin bucket",
					Type: google.GoogleStorageBucketIamBindingResourceType,
					Attrs: &resource.Attributes{
						"role": "storage.admin",
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
					Type: google.GoogleStorageBucketIamBindingResourceType,
					Attrs: &resource.Attributes{
						"role": "storage.admin",
					},
				},
			},
			func(factory *terraform.MockResourceFactory) {

			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := &terraform.MockResourceFactory{}
			if tt.mock != nil {
				tt.mock(factory)
			}

			m := NewGoogleStorageBucketIAMMemberTransformer(factory)
			err := m.Execute(&[]*resource.Resource{}, &tt.resourcesFromState)
			if err != nil {
				t.Fatal(err)
			}
			changelog, err := diff.Diff(tt.expected, tt.resourcesFromState)
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
