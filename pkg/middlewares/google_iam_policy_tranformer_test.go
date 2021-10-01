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

func TestGoogleProjectIAMPolicyTransformer_Execute(t *testing.T) {
	tests := []struct {
		name               string
		resourcesFromState []*resource.Resource
		expected           []*resource.Resource
		mock               func(factory *terraform.MockResourceFactory)
	}{
		{
			"Test that project policy are transformed into bindings",
			[]*resource.Resource{
				{
					Id:    "b/bucket-1",
					Type:  google.GoogleStorageBucketResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:    "b/bucket-2",
					Type:  google.GoogleStorageBucketResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:   "project-1",
					Type: google.GoogleProjectIamPolicyResourceType,
					Attrs: &resource.Attributes{
						"project":     "project-1",
						"id":          "project-1",
						"policy_data": "{\"bindings\":[{\"members\":[\"serviceAccount:driftctl@cloudskiff-dev-martin.iam.gserviceaccount.com\"],\"role\":\"roles/storage.admin\"},{\"members\":[\"user:william.beuil@cloudskiff.com\"],\"role\":\"roles/storage.objectViewer\"}]}",
					},
				},
				{
					Id:   "dctlgstorageprojectiambinding-2",
					Type: google.GoogleProjectIamPolicyResourceType,
					Attrs: &resource.Attributes{
						"project":     "project-2",
						"etag":        "CAU=",
						"id":          "project-2",
						"policy_data": "{\"bindings\":[{\"members\":[\"serviceAccount:driftctl@cloudskiff-dev-martin.iam.gserviceaccount.com\"],\"role\":\"roles/storage.admin\"},{\"members\":[\"user:william.beuil@cloudskiff.com\"],\"role\":\"roles/storage.objectViewer\"}]}",
					},
				},
			},
			[]*resource.Resource{
				{
					Id:    "b/bucket-1",
					Type:  google.GoogleStorageBucketResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:    "b/bucket-2",
					Type:  google.GoogleStorageBucketResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:   "project-1/roles/storage.admin/serviceAccount:driftctl@cloudskiff-dev-martin.iam.gserviceaccount.com",
					Type: google.GoogleProjectIamMemberResourceType,
					Attrs: &resource.Attributes{
						"id":      "project-1/roles/storage.admin/serviceAccount:driftctl@cloudskiff-dev-martin.iam.gserviceaccount.com",
						"role":    "roles/storage.admin",
						"project": "project-1",
						"member":  "serviceAccount:driftctl@cloudskiff-dev-martin.iam.gserviceaccount.com",
					},
				},
				{
					Id:   "project-1/roles/storage.objectViewer/user:william.beuil@cloudskiff.com",
					Type: google.GoogleProjectIamMemberResourceType,
					Attrs: &resource.Attributes{
						"id":      "project-1/roles/storage.objectViewer/user:william.beuil@cloudskiff.com",
						"role":    "roles/storage.objectViewer",
						"project": "project-1",
						"member":  "user:william.beuil@cloudskiff.com",
					},
				},
				{
					Id:   "project-2/roles/storage.admin/serviceAccount:driftctl@cloudskiff-dev-martin.iam.gserviceaccount.com",
					Type: google.GoogleProjectIamMemberResourceType,
					Attrs: &resource.Attributes{
						"id":      "project-2/roles/storage.admin/serviceAccount:driftctl@cloudskiff-dev-martin.iam.gserviceaccount.com",
						"role":    "roles/storage.admin",
						"project": "project-2",
						"member":  "serviceAccount:driftctl@cloudskiff-dev-martin.iam.gserviceaccount.com",
					},
				},
				{
					Id:   "project-2/roles/storage.objectViewer/user:william.beuil@cloudskiff.com",
					Type: google.GoogleProjectIamMemberResourceType,
					Attrs: &resource.Attributes{
						"id":      "project-2/roles/storage.objectViewer/user:william.beuil@cloudskiff.com",
						"role":    "roles/storage.objectViewer",
						"project": "project-2",
						"member":  "user:william.beuil@cloudskiff.com",
					},
				},
			},
			func(factory *terraform.MockResourceFactory) {
				factory.On(
					"CreateAbstractResource", google.GoogleProjectIamMemberResourceType,
					"project-1/roles/storage.admin/serviceAccount:driftctl@cloudskiff-dev-martin.iam.gserviceaccount.com",
					map[string]interface{}{
						"id":      "project-1/roles/storage.admin/serviceAccount:driftctl@cloudskiff-dev-martin.iam.gserviceaccount.com",
						"project": "project-1",
						"role":    "roles/storage.admin",
						"member":  "serviceAccount:driftctl@cloudskiff-dev-martin.iam.gserviceaccount.com",
					}).Return(&resource.Resource{
					Id:   "project-1/roles/storage.admin/serviceAccount:driftctl@cloudskiff-dev-martin.iam.gserviceaccount.com",
					Type: google.GoogleProjectIamMemberResourceType,
					Attrs: &resource.Attributes{
						"id":      "project-1/roles/storage.admin/serviceAccount:driftctl@cloudskiff-dev-martin.iam.gserviceaccount.com",
						"role":    "roles/storage.admin",
						"project": "project-1",
						"member":  "serviceAccount:driftctl@cloudskiff-dev-martin.iam.gserviceaccount.com",
					},
				}).Once()

				factory.On(
					"CreateAbstractResource", google.GoogleProjectIamMemberResourceType,
					"project-1/roles/storage.objectViewer/user:william.beuil@cloudskiff.com",
					map[string]interface{}{
						"id":      "project-1/roles/storage.objectViewer/user:william.beuil@cloudskiff.com",
						"project": "project-1",
						"role":    "roles/storage.objectViewer",
						"member":  "user:william.beuil@cloudskiff.com",
					}).Return(&resource.Resource{
					Id:   "project-1/roles/storage.objectViewer/user:william.beuil@cloudskiff.com",
					Type: google.GoogleProjectIamMemberResourceType,
					Attrs: &resource.Attributes{
						"id":      "project-1/roles/storage.objectViewer/user:william.beuil@cloudskiff.com",
						"role":    "roles/storage.objectViewer",
						"project": "project-1",
						"member":  "user:william.beuil@cloudskiff.com",
					},
				}).Once()

				factory.On(
					"CreateAbstractResource", google.GoogleProjectIamMemberResourceType,
					"project-2/roles/storage.admin/serviceAccount:driftctl@cloudskiff-dev-martin.iam.gserviceaccount.com",
					map[string]interface{}{
						"id":      "project-2/roles/storage.admin/serviceAccount:driftctl@cloudskiff-dev-martin.iam.gserviceaccount.com",
						"project": "project-2",
						"role":    "roles/storage.admin",
						"member":  "serviceAccount:driftctl@cloudskiff-dev-martin.iam.gserviceaccount.com",
					}).Return(&resource.Resource{
					Id:   "project-2/roles/storage.admin/serviceAccount:driftctl@cloudskiff-dev-martin.iam.gserviceaccount.com",
					Type: google.GoogleProjectIamMemberResourceType,
					Attrs: &resource.Attributes{
						"id":      "project-2/roles/storage.admin/serviceAccount:driftctl@cloudskiff-dev-martin.iam.gserviceaccount.com",
						"role":    "roles/storage.admin",
						"project": "project-2",
						"member":  "serviceAccount:driftctl@cloudskiff-dev-martin.iam.gserviceaccount.com",
					},
				}).Once()

				factory.On(
					"CreateAbstractResource", google.GoogleProjectIamMemberResourceType,
					"project-2/roles/storage.objectViewer/user:william.beuil@cloudskiff.com",
					map[string]interface{}{
						"id":      "project-2/roles/storage.objectViewer/user:william.beuil@cloudskiff.com",
						"project": "project-2",
						"role":    "roles/storage.objectViewer",
						"member":  "user:william.beuil@cloudskiff.com",
					}).Return(&resource.Resource{
					Id:   "project-2/roles/storage.objectViewer/user:william.beuil@cloudskiff.com",
					Type: google.GoogleProjectIamMemberResourceType,
					Attrs: &resource.Attributes{
						"id":      "project-2/roles/storage.objectViewer/user:william.beuil@cloudskiff.com",
						"role":    "roles/storage.objectViewer",
						"project": "project-2",
						"member":  "user:william.beuil@cloudskiff.com",
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

			m := NewGoogleIAMPolicyTransformer(factory)
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
						"policy_data": "{\"bindings\":[{\"members\":[\"serviceAccount:driftctl@cloudskiff-dev-martin.iam.gserviceaccount.com\"],\"role\":\"roles/storage.admin\"},{\"members\":[\"user:william.beuil@cloudskiff.com\"],\"role\":\"roles/storage.objectViewer\"}]}",
					},
				},
				{
					Id:   "dctlgstoragebucketiambinding-2",
					Type: google.GoogleStorageBucketIamPolicyResourceType,
					Attrs: &resource.Attributes{
						"bucket":      "b/dctlgstoragebucketiambinding-2",
						"etag":        "CAU=",
						"id":          "b/dctlgstoragebucketiambinding-2",
						"policy_data": "{\"bindings\":[{\"members\":[\"serviceAccount:driftctl@cloudskiff-dev-martin.iam.gserviceaccount.com\"],\"role\":\"roles/storage.admin\"},{\"members\":[\"user:william.beuil@cloudskiff.com\"],\"role\":\"roles/storage.objectViewer\"}]}",
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
					Id:   "b/dctlgstoragebucketiambinding-1/roles/storage.admin/serviceAccount:driftctl@cloudskiff-dev-martin.iam.gserviceaccount.com",
					Type: google.GoogleStorageBucketIamMemberResourceType,
					Attrs: &resource.Attributes{
						"id":     "b/dctlgstoragebucketiambinding-1/roles/storage.admin/serviceAccount:driftctl@cloudskiff-dev-martin.iam.gserviceaccount.com",
						"role":   "roles/storage.admin",
						"bucket": "b/dctlgstoragebucketiambinding-1",
						"member": "serviceAccount:driftctl@cloudskiff-dev-martin.iam.gserviceaccount.com",
					},
				},
				{
					Id:   "b/dctlgstoragebucketiambinding-1/roles/storage.objectViewer/user:william.beuil@cloudskiff.com",
					Type: google.GoogleStorageBucketIamMemberResourceType,
					Attrs: &resource.Attributes{
						"id":     "b/dctlgstoragebucketiambinding-1/roles/storage.objectViewer/user:william.beuil@cloudskiff.com",
						"role":   "roles/storage.objectViewer",
						"bucket": "b/dctlgstoragebucketiambinding-1",
						"member": "user:william.beuil@cloudskiff.com",
					},
				},
				{
					Id:   "b/dctlgstoragebucketiambinding-2/roles/storage.admin/serviceAccount:driftctl@cloudskiff-dev-martin.iam.gserviceaccount.com",
					Type: google.GoogleStorageBucketIamMemberResourceType,
					Attrs: &resource.Attributes{
						"id":     "b/dctlgstoragebucketiambinding-2/roles/storage.admin/serviceAccount:driftctl@cloudskiff-dev-martin.iam.gserviceaccount.com",
						"role":   "roles/storage.admin",
						"bucket": "b/dctlgstoragebucketiambinding-2",
						"member": "serviceAccount:driftctl@cloudskiff-dev-martin.iam.gserviceaccount.com",
					},
				},
				{
					Id:   "b/dctlgstoragebucketiambinding-2/roles/storage.objectViewer/user:william.beuil@cloudskiff.com",
					Type: google.GoogleStorageBucketIamMemberResourceType,
					Attrs: &resource.Attributes{
						"id":     "b/dctlgstoragebucketiambinding-2/roles/storage.objectViewer/user:william.beuil@cloudskiff.com",
						"role":   "roles/storage.objectViewer",
						"bucket": "b/dctlgstoragebucketiambinding-2",
						"member": "user:william.beuil@cloudskiff.com",
					},
				},
			},
			func(factory *terraform.MockResourceFactory) {
				factory.On(
					"CreateAbstractResource", google.GoogleStorageBucketIamMemberResourceType,
					"b/dctlgstoragebucketiambinding-1/roles/storage.admin/serviceAccount:driftctl@cloudskiff-dev-martin.iam.gserviceaccount.com",
					map[string]interface{}{
						"id":     "b/dctlgstoragebucketiambinding-1/roles/storage.admin/serviceAccount:driftctl@cloudskiff-dev-martin.iam.gserviceaccount.com",
						"bucket": "b/dctlgstoragebucketiambinding-1",
						"role":   "roles/storage.admin",
						"member": "serviceAccount:driftctl@cloudskiff-dev-martin.iam.gserviceaccount.com",
					}).Return(&resource.Resource{
					Id:   "b/dctlgstoragebucketiambinding-1/roles/storage.admin/serviceAccount:driftctl@cloudskiff-dev-martin.iam.gserviceaccount.com",
					Type: google.GoogleStorageBucketIamMemberResourceType,
					Attrs: &resource.Attributes{
						"id":     "b/dctlgstoragebucketiambinding-1/roles/storage.admin/serviceAccount:driftctl@cloudskiff-dev-martin.iam.gserviceaccount.com",
						"role":   "roles/storage.admin",
						"bucket": "b/dctlgstoragebucketiambinding-1",
						"member": "serviceAccount:driftctl@cloudskiff-dev-martin.iam.gserviceaccount.com",
					},
				}).Once()

				factory.On(
					"CreateAbstractResource", google.GoogleStorageBucketIamMemberResourceType,
					"b/dctlgstoragebucketiambinding-1/roles/storage.objectViewer/user:william.beuil@cloudskiff.com",
					map[string]interface{}{
						"id":     "b/dctlgstoragebucketiambinding-1/roles/storage.objectViewer/user:william.beuil@cloudskiff.com",
						"bucket": "b/dctlgstoragebucketiambinding-1",
						"role":   "roles/storage.objectViewer",
						"member": "user:william.beuil@cloudskiff.com",
					}).Return(&resource.Resource{
					Id:   "b/dctlgstoragebucketiambinding-1/roles/storage.objectViewer/user:william.beuil@cloudskiff.com",
					Type: google.GoogleStorageBucketIamMemberResourceType,
					Attrs: &resource.Attributes{
						"id":     "b/dctlgstoragebucketiambinding-1/roles/storage.objectViewer/user:william.beuil@cloudskiff.com",
						"role":   "roles/storage.objectViewer",
						"bucket": "b/dctlgstoragebucketiambinding-1",
						"member": "user:william.beuil@cloudskiff.com",
					},
				}).Once()

				factory.On(
					"CreateAbstractResource", google.GoogleStorageBucketIamMemberResourceType,
					"b/dctlgstoragebucketiambinding-2/roles/storage.admin/serviceAccount:driftctl@cloudskiff-dev-martin.iam.gserviceaccount.com",
					map[string]interface{}{
						"id":     "b/dctlgstoragebucketiambinding-2/roles/storage.admin/serviceAccount:driftctl@cloudskiff-dev-martin.iam.gserviceaccount.com",
						"bucket": "b/dctlgstoragebucketiambinding-2",
						"role":   "roles/storage.admin",
						"member": "serviceAccount:driftctl@cloudskiff-dev-martin.iam.gserviceaccount.com",
					}).Return(&resource.Resource{
					Id:   "b/dctlgstoragebucketiambinding-2/roles/storage.admin/serviceAccount:driftctl@cloudskiff-dev-martin.iam.gserviceaccount.com",
					Type: google.GoogleStorageBucketIamMemberResourceType,
					Attrs: &resource.Attributes{
						"id":     "b/dctlgstoragebucketiambinding-2/roles/storage.admin/serviceAccount:driftctl@cloudskiff-dev-martin.iam.gserviceaccount.com",
						"role":   "roles/storage.admin",
						"bucket": "b/dctlgstoragebucketiambinding-2",
						"member": "serviceAccount:driftctl@cloudskiff-dev-martin.iam.gserviceaccount.com",
					},
				}).Once()

				factory.On(
					"CreateAbstractResource", google.GoogleStorageBucketIamMemberResourceType,
					"b/dctlgstoragebucketiambinding-2/roles/storage.objectViewer/user:william.beuil@cloudskiff.com",
					map[string]interface{}{
						"id":     "b/dctlgstoragebucketiambinding-2/roles/storage.objectViewer/user:william.beuil@cloudskiff.com",
						"bucket": "b/dctlgstoragebucketiambinding-2",
						"role":   "roles/storage.objectViewer",
						"member": "user:william.beuil@cloudskiff.com",
					}).Return(&resource.Resource{
					Id:   "b/dctlgstoragebucketiambinding-2/roles/storage.objectViewer/user:william.beuil@cloudskiff.com",
					Type: google.GoogleStorageBucketIamMemberResourceType,
					Attrs: &resource.Attributes{
						"id":     "b/dctlgstoragebucketiambinding-2/roles/storage.objectViewer/user:william.beuil@cloudskiff.com",
						"role":   "roles/storage.objectViewer",
						"bucket": "b/dctlgstoragebucketiambinding-2",
						"member": "user:william.beuil@cloudskiff.com",
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

			m := NewGoogleIAMPolicyTransformer(factory)
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
