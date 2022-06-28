package middlewares

import (
	"github.com/snyk/driftctl/enumeration/terraform"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/r3labs/diff/v2"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/google"
)

func TestGoogleProjectIAMBindingTransformer_Execute(t *testing.T) {
	tests := []struct {
		name               string
		resourcesFromState []*resource.Resource
		expected           []*resource.Resource
		mock               func(factory *terraform.MockResourceFactory)
	}{
		{
			name: "Test that bindings with nil members does not cause any crash",
			resourcesFromState: []*resource.Resource{
				{
					Type: google.GoogleStorageBucketIamBindingResourceType,
					Attrs: &resource.Attributes{
						"bucket":  "hey",
						"role":    "storage.admin",
						"members": nil,
					},
				},
				{
					Type: google.GoogleProjectIamBindingResourceType,
					Attrs: &resource.Attributes{
						"project": "coucou",
						"role":    "storage.admin",
						"members": nil,
					},
				},
			},
			expected: []*resource.Resource{},
			mock:     nil,
		},
		{
			"Test that project bindings are transformed into member",
			[]*resource.Resource{
				{
					Id:    "fake",
					Type:  google.GoogleStorageBucketResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:   "admin project",
					Type: google.GoogleProjectIamMemberResourceType,
					Attrs: &resource.Attributes{
						"project": "coucou",
						"role":    "storage.admin",
						"member":  "user:elie@cloudskiff.com",
					},
				},
				{
					Id:   "proj/admin",
					Type: google.GoogleProjectIamBindingResourceType,
					Attrs: &resource.Attributes{
						"role":    "storage.admin",
						"project": "proj",
						"members": []interface{}{
							"user:elie@cloudskiff.com",
							"user:william@cloudskiff.com",
						},
					},
				},

				{
					Id:   "proj/viewer",
					Type: google.GoogleProjectIamBindingResourceType,
					Attrs: &resource.Attributes{
						"role":    "storage.viewer",
						"project": "proj",
						"members": []interface{}{
							"user:william@cloudskiff.com",
						},
					},
				},
				{
					Id:   "proj2/viewer",
					Type: google.GoogleProjectIamBindingResourceType,
					Attrs: &resource.Attributes{
						"role":    "storage.viewer",
						"project": "proj2",
						"members": []interface{}{
							"user:william@cloudskiff.com",
						},
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
					Id:   "admin project",
					Type: google.GoogleProjectIamMemberResourceType,
					Attrs: &resource.Attributes{
						"project": "coucou",
						"role":    "storage.admin",
						"member":  "user:elie@cloudskiff.com",
					},
				},
				{
					Id:   "proj/storage.admin/user:elie@cloudskiff.com",
					Type: google.GoogleProjectIamMemberResourceType,
					Attrs: &resource.Attributes{
						"role":    "storage.admin",
						"project": "proj",
						"member":  "user:elie@cloudskiff.com",
					},
				},
				{
					Id:   "proj/storage.admin/user:william@cloudskiff.com",
					Type: google.GoogleProjectIamMemberResourceType,
					Attrs: &resource.Attributes{
						"role":    "storage.admin",
						"project": "proj",
						"member":  "user:william@cloudskiff.com",
					},
				},
				{
					Id:   "proj/storage.viewer/user:william@cloudskiff.com",
					Type: google.GoogleProjectIamMemberResourceType,
					Attrs: &resource.Attributes{
						"role":    "storage.viewer",
						"project": "proj",
						"member":  "user:william@cloudskiff.com",
					},
				},
				{
					Id:   "proj2/storage.viewer/user:william@cloudskiff.com",
					Type: google.GoogleProjectIamMemberResourceType,
					Attrs: &resource.Attributes{
						"role":    "storage.viewer",
						"project": "proj2",
						"member":  "user:william@cloudskiff.com",
					},
				},
			},
			func(factory *terraform.MockResourceFactory) {
				factory.On(
					"CreateAbstractResource", google.GoogleProjectIamMemberResourceType,
					"proj/storage.admin/user:elie@cloudskiff.com",
					map[string]interface{}{
						"id":      "proj/storage.admin/user:elie@cloudskiff.com",
						"project": "proj",
						"role":    "storage.admin",
						"member":  "user:elie@cloudskiff.com",
					}).Return(&resource.Resource{
					Id:   "proj/storage.admin/user:elie@cloudskiff.com",
					Type: google.GoogleProjectIamMemberResourceType,
					Attrs: &resource.Attributes{
						"role":    "storage.admin",
						"project": "proj",
						"member":  "user:elie@cloudskiff.com",
					},
				}).Once()

				factory.On(
					"CreateAbstractResource", google.GoogleProjectIamMemberResourceType,
					"proj/storage.admin/user:william@cloudskiff.com",
					map[string]interface{}{
						"id":      "proj/storage.admin/user:william@cloudskiff.com",
						"project": "proj",
						"role":    "storage.admin",
						"member":  "user:william@cloudskiff.com",
					}).Return(&resource.Resource{
					Id:   "proj/storage.admin/user:william@cloudskiff.com",
					Type: google.GoogleProjectIamMemberResourceType,
					Attrs: &resource.Attributes{
						"role":    "storage.admin",
						"project": "proj",
						"member":  "user:william@cloudskiff.com",
					},
				}).Once()

				factory.On(
					"CreateAbstractResource", google.GoogleProjectIamMemberResourceType,
					"proj/storage.viewer/user:william@cloudskiff.com",
					map[string]interface{}{
						"id":      "proj/storage.viewer/user:william@cloudskiff.com",
						"project": "proj",
						"role":    "storage.viewer",
						"member":  "user:william@cloudskiff.com",
					}).Return(&resource.Resource{
					Id:   "proj/storage.viewer/user:william@cloudskiff.com",
					Type: google.GoogleProjectIamMemberResourceType,
					Attrs: &resource.Attributes{
						"role":    "storage.viewer",
						"project": "proj",
						"member":  "user:william@cloudskiff.com",
					},
				}).Once()

				factory.On(
					"CreateAbstractResource", google.GoogleProjectIamMemberResourceType,
					"proj2/storage.viewer/user:william@cloudskiff.com",
					map[string]interface{}{
						"id":      "proj2/storage.viewer/user:william@cloudskiff.com",
						"project": "proj2",
						"role":    "storage.viewer",
						"member":  "user:william@cloudskiff.com",
					}).Return(&resource.Resource{
					Id:   "proj2/storage.viewer/user:william@cloudskiff.com",
					Type: google.GoogleProjectIamMemberResourceType,
					Attrs: &resource.Attributes{
						"role":    "storage.viewer",
						"project": "proj2",
						"member":  "user:william@cloudskiff.com",
					},
				}).Once()
			},
		},
		{
			"test that everything is fine when there is no bindings",
			[]*resource.Resource{
				{
					Id:    "fake",
					Type:  google.GoogleStorageBucketResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:   "admin project",
					Type: google.GoogleProjectIamMemberResourceType,
					Attrs: &resource.Attributes{
						"project": "coucou",
						"role":    "storage.admin",
						"member":  "user:elie@cloudskiff.com",
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
					Id:   "admin project",
					Type: google.GoogleProjectIamMemberResourceType,
					Attrs: &resource.Attributes{
						"project": "coucou",
						"role":    "storage.admin",
						"member":  "user:elie@cloudskiff.com",
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

			m := NewGoogleIAMBindingTransformer(factory)
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

func TestGoogleBucketIAMBindingTransformer_Execute(t *testing.T) {
	tests := []struct {
		name               string
		resourcesFromState []*resource.Resource
		expected           []*resource.Resource
		mock               func(factory *terraform.MockResourceFactory)
	}{
		{
			"Test that bucket bindings are transformed into member",
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
						"bucket": "coucou",
						"role":   "storage.admin",
						"member": "user:elie@cloudskiff.com",
					},
				},
				{
					Id:   "b/bucket/admin",
					Type: google.GoogleStorageBucketIamBindingResourceType,
					Attrs: &resource.Attributes{
						"role":   "storage.admin",
						"bucket": "b/bucket",
						"members": []interface{}{
							"user:elie@cloudskiff.com",
							"user:william@cloudskiff.com",
						},
					},
				},

				{
					Id:   "b/bucket/viewer",
					Type: google.GoogleStorageBucketIamBindingResourceType,
					Attrs: &resource.Attributes{
						"role":   "storage.viewer",
						"bucket": "b/bucket",
						"members": []interface{}{
							"user:william@cloudskiff.com",
						},
					},
				},
				{
					Id:   "b/bucket2/viewer",
					Type: google.GoogleStorageBucketIamBindingResourceType,
					Attrs: &resource.Attributes{
						"role":   "storage.viewer",
						"bucket": "b/bucket2",
						"members": []interface{}{
							"user:william@cloudskiff.com",
						},
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
						"bucket": "coucou",
						"role":   "storage.admin",
						"member": "user:elie@cloudskiff.com",
					},
				},
				{
					Id:   "b/bucket/storage.admin/user:elie@cloudskiff.com",
					Type: google.GoogleStorageBucketIamMemberResourceType,
					Attrs: &resource.Attributes{
						"role":   "storage.admin",
						"bucket": "b/bucket",
						"member": "user:elie@cloudskiff.com",
					},
				},
				{
					Id:   "b/bucket/storage.admin/user:william@cloudskiff.com",
					Type: google.GoogleStorageBucketIamMemberResourceType,
					Attrs: &resource.Attributes{
						"role":   "storage.admin",
						"bucket": "b/bucket",
						"member": "user:william@cloudskiff.com",
					},
				},
				{
					Id:   "b/bucket/storage.viewer/user:william@cloudskiff.com",
					Type: google.GoogleStorageBucketIamMemberResourceType,
					Attrs: &resource.Attributes{
						"role":   "storage.viewer",
						"bucket": "b/bucket",
						"member": "user:william@cloudskiff.com",
					},
				},
				{
					Id:   "b/bucket2/storage.viewer/user:william@cloudskiff.com",
					Type: google.GoogleStorageBucketIamMemberResourceType,
					Attrs: &resource.Attributes{
						"role":   "storage.viewer",
						"bucket": "b/bucket2",
						"member": "user:william@cloudskiff.com",
					},
				},
			},
			func(factory *terraform.MockResourceFactory) {
				factory.On(
					"CreateAbstractResource", google.GoogleStorageBucketIamMemberResourceType,
					"b/bucket/storage.admin/user:elie@cloudskiff.com",
					map[string]interface{}{
						"id":     "b/bucket/storage.admin/user:elie@cloudskiff.com",
						"bucket": "b/bucket",
						"role":   "storage.admin",
						"member": "user:elie@cloudskiff.com",
					}).Return(&resource.Resource{
					Id:   "b/bucket/storage.admin/user:elie@cloudskiff.com",
					Type: google.GoogleStorageBucketIamMemberResourceType,
					Attrs: &resource.Attributes{
						"role":   "storage.admin",
						"bucket": "b/bucket",
						"member": "user:elie@cloudskiff.com",
					},
				}).Once()

				factory.On(
					"CreateAbstractResource", google.GoogleStorageBucketIamMemberResourceType,
					"b/bucket/storage.admin/user:william@cloudskiff.com",
					map[string]interface{}{
						"id":     "b/bucket/storage.admin/user:william@cloudskiff.com",
						"bucket": "b/bucket",
						"role":   "storage.admin",
						"member": "user:william@cloudskiff.com",
					}).Return(&resource.Resource{
					Id:   "b/bucket/storage.admin/user:william@cloudskiff.com",
					Type: google.GoogleStorageBucketIamMemberResourceType,
					Attrs: &resource.Attributes{
						"role":   "storage.admin",
						"bucket": "b/bucket",
						"member": "user:william@cloudskiff.com",
					},
				}).Once()

				factory.On(
					"CreateAbstractResource", google.GoogleStorageBucketIamMemberResourceType,
					"b/bucket/storage.viewer/user:william@cloudskiff.com",
					map[string]interface{}{
						"id":     "b/bucket/storage.viewer/user:william@cloudskiff.com",
						"bucket": "b/bucket",
						"role":   "storage.viewer",
						"member": "user:william@cloudskiff.com",
					}).Return(&resource.Resource{
					Id:   "b/bucket/storage.viewer/user:william@cloudskiff.com",
					Type: google.GoogleStorageBucketIamMemberResourceType,
					Attrs: &resource.Attributes{
						"role":   "storage.viewer",
						"bucket": "b/bucket",
						"member": "user:william@cloudskiff.com",
					},
				}).Once()

				factory.On(
					"CreateAbstractResource", google.GoogleStorageBucketIamMemberResourceType,
					"b/bucket2/storage.viewer/user:william@cloudskiff.com",
					map[string]interface{}{
						"id":     "b/bucket2/storage.viewer/user:william@cloudskiff.com",
						"bucket": "b/bucket2",
						"role":   "storage.viewer",
						"member": "user:william@cloudskiff.com",
					}).Return(&resource.Resource{
					Id:   "b/bucket2/storage.viewer/user:william@cloudskiff.com",
					Type: google.GoogleStorageBucketIamMemberResourceType,
					Attrs: &resource.Attributes{
						"role":   "storage.viewer",
						"bucket": "b/bucket2",
						"member": "user:william@cloudskiff.com",
					},
				}).Once()
			},
		},
		{
			"test that everything is fine when there is no bindings",
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
						"bucket": "coucou",
						"role":   "storage.admin",
						"member": "user:elie@cloudskiff.com",
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
						"bucket": "coucou",
						"role":   "storage.admin",
						"member": "user:elie@cloudskiff.com",
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

			m := NewGoogleIAMBindingTransformer(factory)
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
