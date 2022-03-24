package middlewares

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/r3labs/diff/v2"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/google"
)

func TestGoogleComputeInstanceGroupManagerExpander_Execute(t *testing.T) {
	tests := []struct {
		name               string
		remoteResources    []*resource.Resource
		resourcesFromState []*resource.Resource
		expected           []*resource.Resource
	}{
		{
			name: "test that we import compute instance group in the state",
			remoteResources: []*resource.Resource{
				{
					Id:   "appserver-igm",
					Type: google.GoogleComputeInstanceGroupManagerResourceType,
					Attrs: &resource.Attributes{
						"name": "appserver-igm",
					},
				},
				{
					Id:   "appserver-igm",
					Type: google.GoogleComputeInstanceGroupResourceType,
					Attrs: &resource.Attributes{
						"name": "appserver-igm",
					},
				},
			},
			resourcesFromState: []*resource.Resource{
				{
					Id:   "appserver-igm",
					Type: google.GoogleComputeInstanceGroupManagerResourceType,
					Attrs: &resource.Attributes{
						"name": "appserver-igm",
					},
				},
				{
					Id:    "fake",
					Type:  google.GoogleComputeInstanceResourceType,
					Attrs: &resource.Attributes{},
				},
			},
			expected: []*resource.Resource{
				{
					Id:   "appserver-igm",
					Type: google.GoogleComputeInstanceGroupManagerResourceType,
					Attrs: &resource.Attributes{
						"name": "appserver-igm",
					},
				},
				{
					Id:   "appserver-igm",
					Type: google.GoogleComputeInstanceGroupResourceType,
					Attrs: &resource.Attributes{
						"name": "appserver-igm",
					},
				},
				{
					Id:    "fake",
					Type:  google.GoogleComputeInstanceResourceType,
					Attrs: &resource.Attributes{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewGoogleComputeInstanceGroupManagerInstances()
			err := m.Execute(&tt.remoteResources, &tt.resourcesFromState)
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
