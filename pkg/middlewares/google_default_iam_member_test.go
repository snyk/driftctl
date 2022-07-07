package middlewares

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/r3labs/diff/v2"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/google"
)

func TestGoogleDefaultIAMMember_Execute(t *testing.T) {
	tests := []struct {
		name               string
		remoteResources    []*resource.Resource
		resourcesFromState []*resource.Resource
		expected           []*resource.Resource
	}{
		{
			name: "test that we ignore only default service account",
			remoteResources: []*resource.Resource{
				{
					Id:    "fake",
					Type:  google.GoogleStorageBucketResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:   "user",
					Type: google.GoogleProjectIamMemberResourceType,
					Attrs: &resource.Attributes{
						"role":    "role",
						"member":  "user:test@user.com",
						"project": "project",
					},
				},
				{
					Id:   "serviceaccount",
					Type: google.GoogleProjectIamMemberResourceType,
					Attrs: &resource.Attributes{
						"role":    "role",
						"member":  "serviceAccount:test@project.iam.gserviceaccount.com",
						"project": "project",
					},
				},
				{
					Id:   "default",
					Type: google.GoogleProjectIamMemberResourceType,
					Attrs: &resource.Attributes{
						"role":    "role",
						"member":  "serviceAccount:cloudskiff-dev-martin@appspot.gserviceaccount.com ",
						"project": "project",
					},
				},
			},
			resourcesFromState: []*resource.Resource{},
			expected: []*resource.Resource{
				{
					Id:    "fake",
					Type:  google.GoogleStorageBucketResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:   "user",
					Type: google.GoogleProjectIamMemberResourceType,
					Attrs: &resource.Attributes{
						"role":    "role",
						"member":  "user:test@user.com",
						"project": "project",
					},
				},
				{
					Id:   "serviceaccount",
					Type: google.GoogleProjectIamMemberResourceType,
					Attrs: &resource.Attributes{
						"role":    "role",
						"member":  "serviceAccount:test@project.iam.gserviceaccount.com",
						"project": "project",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewGoogleDefaultIAMMember()
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
