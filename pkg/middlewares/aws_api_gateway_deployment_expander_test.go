package middlewares

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
	"github.com/snyk/driftctl/pkg/terraform"

	"github.com/r3labs/diff/v2"
)

func TestAwsApiGatewayDeploymentExpander_Execute(t *testing.T) {
	tests := []struct {
		name               string
		resourcesFromState []*resource.Resource
		mocks              func(*terraform.MockResourceFactory)
		expected           []*resource.Resource
	}{
		{
			name: "no stages created from deployment state resources",
			resourcesFromState: []*resource.Resource{
				{
					Id:   "foo",
					Type: aws.AwsApiGatewayDeploymentResourceType,
					Attrs: &resource.Attributes{
						"rest_api_id": "api",
					},
				},
				{
					Id:   "bar",
					Type: aws.AwsApiGatewayDeploymentResourceType,
					Attrs: &resource.Attributes{
						"stage_name":  "",
						"rest_api_id": "api",
					},
				},
				{
					Id:   "ags-api-baz",
					Type: aws.AwsApiGatewayStageResourceType,
					Attrs: &resource.Attributes{
						"stage_name": "baz",
					},
				},
			},
			expected: []*resource.Resource{
				{
					Id:   "ags-api-baz",
					Type: aws.AwsApiGatewayStageResourceType,
					Attrs: &resource.Attributes{
						"stage_name": "baz",
					},
				},
			},
		},
		{
			name: "stages created from deployment state resources",
			mocks: func(factory *terraform.MockResourceFactory) {
				factory.On(
					"CreateAbstractResource",
					aws.AwsApiGatewayStageResourceType,
					"ags-api-foo",
					map[string]interface{}{},
				).Once().Return(&resource.Resource{
					Id:   "ags-api-foo",
					Type: aws.AwsApiGatewayStageResourceType,
				})
			},
			resourcesFromState: []*resource.Resource{
				{
					Id:   "foo",
					Type: aws.AwsApiGatewayDeploymentResourceType,
					Attrs: &resource.Attributes{
						"stage_name":  "foo",
						"rest_api_id": "api",
					},
				},
				{
					Id:   "ags-api-baz",
					Type: aws.AwsApiGatewayStageResourceType,
					Attrs: &resource.Attributes{
						"stage_name": "baz",
					},
				},
			},
			expected: []*resource.Resource{
				{
					Id:   "ags-api-baz",
					Type: aws.AwsApiGatewayStageResourceType,
					Attrs: &resource.Attributes{
						"stage_name": "baz",
					},
				},
				{
					Id:   "ags-api-foo",
					Type: aws.AwsApiGatewayStageResourceType,
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

			m := NewAwsApiGatewayDeploymentExpander(factory)
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
