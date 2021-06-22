package middlewares

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/r3labs/diff/v2"
)

func TestTagsAllManager_Execute(t *testing.T) {
	tests := []struct {
		name               string
		remoteResources    *[]resource.Resource
		resourcesFromState *[]resource.Resource
		wantErr            bool
	}{
		{
			name: "With multiple resources that are tags_all compatible",
			remoteResources: &[]resource.Resource{
				&resource.AbstractResource{
					Id:   "dummy-instance",
					Type: "aws_instance",
					Attrs: &resource.Attributes{
						"tags": map[string]interface{}{
							"Name":      "toto",
							"Terraform": "true",
						},
						"tags_all": map[string]interface{}{
							"Name":      "toto",
							"Terraform": "true",
						},
					},
				},
				&resource.AbstractResource{
					Id:   "dummy-ebs-volume",
					Type: "aws_ebs_volume",
					Attrs: &resource.Attributes{
						"tags": map[string]interface{}{
							"Name":      "tata",
							"Terraform": "true",
						},
						"tags_all": map[string]interface{}{
							"Name":      "tata",
							"Terraform": "true",
						},
					},
				},
			},
			resourcesFromState: &[]resource.Resource{
				&resource.AbstractResource{
					Id:   "dummy-instance",
					Type: "aws_instance",
					Attrs: &resource.Attributes{
						"tags": map[string]interface{}{
							"Name": "toto",
						},
						"tags_all": map[string]interface{}{
							"Name":      "toto",
							"Terraform": "true",
						},
					},
				},
				&resource.AbstractResource{
					Id:   "dummy-ebs-volume",
					Type: "aws_ebs_volume",
					Attrs: &resource.Attributes{
						"tags": map[string]interface{}{
							"Name": "tata",
						},
						"tags_all": map[string]interface{}{
							"Name":      "tata",
							"Terraform": "true",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := NewTagsAllManager()
			if err := a.Execute(tt.remoteResources, tt.resourcesFromState); (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
			changelog, err := diff.Diff(tt.resourcesFromState, tt.remoteResources)
			if err != nil {
				t.Error(err)
			}
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s got = %v, want %v", strings.Join(change.Path, "."), awsutil.Prettify(change.From), awsutil.Prettify(change.To))
				}
			}
		})
	}
}
