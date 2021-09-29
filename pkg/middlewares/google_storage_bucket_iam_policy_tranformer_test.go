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

func TestGoogleBucketIAMPolicyTransformer_Execute(t *testing.T) {
	tests := []struct {
		name               string
		resourcesFromState []*resource.Resource
		expected           []*resource.Resource
		mock               func(factory *terraform.MockResourceFactory)
	}{
		{
			"Test that bucket policy are transformed into bindings",
			[]*resource.Resource{
				{
					Id:    "b/dctlgstoragebucketiambinding-1",
					Type:  google.GoogleStorageBucketResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:    "b/dctlgstoragebucketiambinding-2",
					Type:  google.GoogleStorageBucketResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:   "b/dctlgstoragebucketiambinding-1",
					Type: google.GoogleStorageBucketIamPolicyResourceType,
					Attrs: &resource.Attributes{
						"bucket":      "b/dctlgstoragebucketiambinding-1",
						"id":          "b/dctlgstoragebucketiambinding-1",
						"policy_data": "{\"bindings\":[{\"members\":[\"serviceAccount:driftctl@cloudskiff-dev-martin.iam.gserviceaccount.com\",\"user:elie.charra@cloudskiff.com\",\"user:martin.guibert@cloudskiff.com\"],\"role\":\"roles/storage.admin\"},{\"members\":[\"user:william.beuil@cloudskiff.com\"],\"role\":\"roles/storage.objectViewer\"}]}",
					},
				},
				{
					Id:   "dctlgstoragebucketiambinding-2",
					Type: google.GoogleStorageBucketIamPolicyResourceType,
					Attrs: &resource.Attributes{
						"bucket":      "b/dctlgstoragebucketiambinding-2",
						"etag":        "CAU=",
						"id":          "b/dctlgstoragebucketiambinding-2",
						"policy_data": "{\"bindings\":[{\"members\":[\"serviceAccount:driftctl@cloudskiff-dev-martin.iam.gserviceaccount.com\",\"user:elie.charra@cloudskiff.com\",\"user:martin.guibert@cloudskiff.com\"],\"role\":\"roles/storage.admin\"},{\"members\":[\"user:william.beuil@cloudskiff.com\"],\"role\":\"roles/storage.objectViewer\"}]}",
					},
				},
			},
			[]*resource.Resource{
				{
					Id:    "b/dctlgstoragebucketiambinding-1",
					Type:  google.GoogleStorageBucketResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:    "b/dctlgstoragebucketiambinding-2",
					Type:  google.GoogleStorageBucketResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:   "b/dctlgstoragebucketiambinding-1/roles/storage.admin",
					Type: google.GoogleStorageBucketIamBindingResourceType,
					Attrs: &resource.Attributes{
						"id":     "b/dctlgstoragebucketiambinding-1/roles/storage.admin",
						"role":   "roles/storage.admin",
						"bucket": "b/dctlgstoragebucketiambinding-1",
						"members": []string{
							"user:elie.charra@cloudskiff.com",
							"user:martin.guibert@cloudskiff.com",
							"serviceAccount:driftctl@cloudskiff-dev-martin.iam.gserviceaccount.com",
						},
					},
				},
				{
					Id:   "b/dctlgstoragebucketiambinding-1/roles/storage.objectViewer",
					Type: google.GoogleStorageBucketIamBindingResourceType,
					Attrs: &resource.Attributes{
						"id":     "b/dctlgstoragebucketiambinding-1/roles/storage.objectViewer",
						"role":   "roles/storage.objectViewer",
						"bucket": "b/dctlgstoragebucketiambinding-1",
						"members": []string{
							"user:william.beuil@cloudskiff.com",
						},
					},
				},
				{
					Id:   "b/dctlgstoragebucketiambinding-2/roles/storage.admin",
					Type: google.GoogleStorageBucketIamBindingResourceType,
					Attrs: &resource.Attributes{
						"id":     "b/dctlgstoragebucketiambinding-2/roles/storage.admin",
						"role":   "roles/storage.admin",
						"bucket": "b/dctlgstoragebucketiambinding-2",
						"members": []string{
							"user:elie.charra@cloudskiff.com",
							"user:martin.guibert@cloudskiff.com",
							"serviceAccount:driftctl@cloudskiff-dev-martin.iam.gserviceaccount.com",
						},
					},
				},
				{
					Id:   "b/dctlgstoragebucketiambinding-2/roles/storage.objectViewer",
					Type: google.GoogleStorageBucketIamBindingResourceType,
					Attrs: &resource.Attributes{
						"id":     "b/dctlgstoragebucketiambinding-2/roles/storage.objectViewer",
						"role":   "roles/storage.objectViewer",
						"bucket": "b/dctlgstoragebucketiambinding-2",
						"members": []string{
							"user:william.beuil@cloudskiff.com",
						},
					},
				},
			},
			func(factory *terraform.MockResourceFactory) {
				factory.On(
					"CreateAbstractResource", google.GoogleStorageBucketIamBindingResourceType,
					"b/dctlgstoragebucketiambinding-1/roles/storage.admin",
					map[string]interface{}{
						"id":     "b/dctlgstoragebucketiambinding-1/roles/storage.admin",
						"bucket": "b/dctlgstoragebucketiambinding-1",
						"role":   "roles/storage.admin",
						"members": []string{
							"serviceAccount:driftctl@cloudskiff-dev-martin.iam.gserviceaccount.com",
							"user:elie.charra@cloudskiff.com",
							"user:martin.guibert@cloudskiff.com",
						},
					}).Return(&resource.Resource{
					Id:   "b/dctlgstoragebucketiambinding-1/roles/storage.admin",
					Type: google.GoogleStorageBucketIamBindingResourceType,
					Attrs: &resource.Attributes{
						"id":     "b/dctlgstoragebucketiambinding-1/roles/storage.admin",
						"role":   "roles/storage.admin",
						"bucket": "b/dctlgstoragebucketiambinding-1",
						"members": []string{
							"serviceAccount:driftctl@cloudskiff-dev-martin.iam.gserviceaccount.com",
							"user:elie.charra@cloudskiff.com",
							"user:martin.guibert@cloudskiff.com",
						},
					},
				}).Once()

				factory.On(
					"CreateAbstractResource", google.GoogleStorageBucketIamBindingResourceType,
					"b/dctlgstoragebucketiambinding-1/roles/storage.objectViewer",
					map[string]interface{}{
						"id":     "b/dctlgstoragebucketiambinding-1/roles/storage.objectViewer",
						"bucket": "b/dctlgstoragebucketiambinding-1",
						"role":   "roles/storage.objectViewer",
						"members": []string{
							"user:william.beuil@cloudskiff.com",
						},
					}).Return(&resource.Resource{
					Id:   "b/dctlgstoragebucketiambinding-1/roles/storage.objectViewer",
					Type: google.GoogleStorageBucketIamBindingResourceType,
					Attrs: &resource.Attributes{
						"id":     "b/dctlgstoragebucketiambinding-1/roles/storage.objectViewer",
						"role":   "roles/storage.objectViewer",
						"bucket": "b/dctlgstoragebucketiambinding-1",
						"members": []string{
							"user:william.beuil@cloudskiff.com",
						},
					},
				}).Once()

				factory.On(
					"CreateAbstractResource", google.GoogleStorageBucketIamBindingResourceType,
					"b/dctlgstoragebucketiambinding-2/roles/storage.admin",
					map[string]interface{}{
						"id":     "b/dctlgstoragebucketiambinding-2/roles/storage.admin",
						"bucket": "b/dctlgstoragebucketiambinding-2",
						"role":   "roles/storage.admin",
						"members": []string{
							"serviceAccount:driftctl@cloudskiff-dev-martin.iam.gserviceaccount.com",
							"user:elie.charra@cloudskiff.com",
							"user:martin.guibert@cloudskiff.com",
						},
					}).Return(&resource.Resource{
					Id:   "b/dctlgstoragebucketiambinding-2/roles/storage.admin",
					Type: google.GoogleStorageBucketIamBindingResourceType,
					Attrs: &resource.Attributes{
						"id":     "b/dctlgstoragebucketiambinding-2/roles/storage.admin",
						"role":   "roles/storage.admin",
						"bucket": "b/dctlgstoragebucketiambinding-2",
						"members": []string{
							"serviceAccount:driftctl@cloudskiff-dev-martin.iam.gserviceaccount.com",
							"user:elie.charra@cloudskiff.com",
							"user:martin.guibert@cloudskiff.com",
						},
					},
				}).Once()

				factory.On(
					"CreateAbstractResource", google.GoogleStorageBucketIamBindingResourceType,
					"b/dctlgstoragebucketiambinding-2/roles/storage.objectViewer",
					map[string]interface{}{
						"id":     "b/dctlgstoragebucketiambinding-2/roles/storage.objectViewer",
						"bucket": "b/dctlgstoragebucketiambinding-2",
						"role":   "roles/storage.objectViewer",
						"members": []string{
							"user:william.beuil@cloudskiff.com",
						},
					}).Return(&resource.Resource{
					Id:   "b/dctlgstoragebucketiambinding-2/roles/storage.objectViewer",
					Type: google.GoogleStorageBucketIamBindingResourceType,
					Attrs: &resource.Attributes{
						"id":     "b/dctlgstoragebucketiambinding-2/roles/storage.objectViewer",
						"role":   "roles/storage.objectViewer",
						"bucket": "b/dctlgstoragebucketiambinding-2",
						"members": []string{
							"user:william.beuil@cloudskiff.com",
						},
					},
				}).Once()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := &terraform.MockResourceFactory{}
			if tt.mock != nil {
				tt.mock(factory)
			}

			m := NewGoogleStorageBucketIAMPolicyTransformer(factory)
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
