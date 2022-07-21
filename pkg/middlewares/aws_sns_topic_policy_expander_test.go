package middlewares

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/r3labs/diff/v2"
	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
	awsresource "github.com/snyk/driftctl/pkg/resource/aws"
	testresource "github.com/snyk/driftctl/test/resource"
	"github.com/stretchr/testify/mock"
)

func TestAwsSNSTopicPolicyExpander_Execute(t *testing.T) {
	tests := []struct {
		name               string
		resourcesFromState *[]*resource.Resource
		expected           *[]*resource.Resource
		mock               func(factory *dctlresource.MockResourceFactory)
		wantErr            bool
	}{
		{
			name: "Inline policy no attached policy",
			resourcesFromState: &[]*resource.Resource{
				{
					Id:   "ID",
					Type: awsresource.AwsSnsTopicResourceType,
					Attrs: &resource.Attributes{
						"arn":    "arn",
						"id":     "ID",
						"policy": "{\"policy\":\"coucou\"}",
					},
				},
			},
			expected: &[]*resource.Resource{
				{
					Id:   "ID",
					Type: awsresource.AwsSnsTopicResourceType,
					Attrs: &resource.Attributes{
						"arn": "arn",
						"id":  "ID",
					},
				},
				{
					Id:   "ID",
					Type: awsresource.AwsSnsTopicPolicyResourceType,
					Attrs: &resource.Attributes{
						"arn":    "arn",
						"id":     "ID",
						"policy": "{\"policy\":\"coucou\"}",
					},
				},
			},
			mock: func(factory *dctlresource.MockResourceFactory) {
				factory.On("CreateAbstractResource", awsresource.AwsSnsTopicPolicyResourceType, "ID", map[string]interface{}{
					"arn":    "arn",
					"id":     "ID",
					"policy": "{\"policy\":\"coucou\"}",
				}).Once().Return(&resource.Resource{
					Id:   "ID",
					Type: awsresource.AwsSnsTopicPolicyResourceType,
					Attrs: &resource.Attributes{
						"arn":    "arn",
						"id":     "ID",
						"policy": "{\"policy\":\"coucou\"}",
					},
				}, nil)
			},
			wantErr: false,
		},
		{
			name: "No inline policy, attached policy",
			resourcesFromState: &[]*resource.Resource{
				{
					Id:   "ID",
					Type: awsresource.AwsSnsTopicResourceType,
					Attrs: &resource.Attributes{
						"arn": "arn",
						"id":  "ID",
					},
				},
				{
					Id:   "ID",
					Type: awsresource.AwsSnsTopicPolicyResourceType,
					Attrs: &resource.Attributes{
						"arn":    "arn",
						"id":     "ID",
						"policy": "{\"policy\":\"coucou\"}",
					},
				},
			},
			expected: &[]*resource.Resource{
				{
					Id:   "ID",
					Type: awsresource.AwsSnsTopicResourceType,
					Attrs: &resource.Attributes{
						"arn": "arn",
						"id":  "ID",
					},
				},
				{
					Id:   "ID",
					Type: awsresource.AwsSnsTopicPolicyResourceType,
					Attrs: &resource.Attributes{
						"arn":    "arn",
						"id":     "ID",
						"policy": "{\"policy\":\"coucou\"}",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "inline policy and dup attached policy",
			resourcesFromState: &[]*resource.Resource{
				{
					Id:   "ID",
					Type: awsresource.AwsSnsTopicResourceType,
					Attrs: &resource.Attributes{
						"arn":    "arn",
						"id":     "ID",
						"policy": "{\"policy\":\"coucou\"}",
					},
				},
				{
					Id:   "ID",
					Type: awsresource.AwsSnsTopicPolicyResourceType,
					Attrs: &resource.Attributes{
						"arn":    "arn",
						"id":     "ID",
						"policy": "{\"policy\":\"coucou\"}",
					},
				},
			},
			expected: &[]*resource.Resource{
				{
					Id:   "ID",
					Type: awsresource.AwsSnsTopicResourceType,
					Attrs: &resource.Attributes{
						"arn": "arn",
						"id":  "ID",
					},
				},
				{
					Id:   "ID",
					Type: awsresource.AwsSnsTopicPolicyResourceType,
					Attrs: &resource.Attributes{
						"arn":    "arn",
						"id":     "ID",
						"policy": "{\"policy\":\"coucou\"}",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "inline policy and attached policy",
			resourcesFromState: &[]*resource.Resource{
				{
					Id:   "ID",
					Type: awsresource.AwsSnsTopicResourceType,
					Attrs: &resource.Attributes{
						"arn":    "arn",
						"id":     "ID",
						"policy": "{\"policy\":\"coucou\"}",
					},
				},
				{
					Id:   "ID2",
					Type: awsresource.AwsSnsTopicPolicyResourceType,
					Attrs: &resource.Attributes{
						"arn":    "arn2",
						"id":     "ID2",
						"policy": "{\"policy\":\"coucou2\"}",
					},
				},
			},
			expected: &[]*resource.Resource{
				{
					Id:   "ID",
					Type: awsresource.AwsSnsTopicResourceType,
					Attrs: &resource.Attributes{
						"arn": "arn",
						"id":  "ID",
					},
				},
				{
					Id:   "ID",
					Type: awsresource.AwsSnsTopicPolicyResourceType,
					Attrs: &resource.Attributes{
						"arn":    "arn",
						"id":     "ID",
						"policy": "{\"policy\":\"coucou\"}",
					},
				},
				{
					Id:   "ID2",
					Type: awsresource.AwsSnsTopicPolicyResourceType,
					Attrs: &resource.Attributes{
						"arn":    "arn2",
						"id":     "ID2",
						"policy": "{\"policy\":\"coucou2\"}",
					},
				},
			},
			mock: func(factory *dctlresource.MockResourceFactory) {
				factory.On("CreateAbstractResource", awsresource.AwsSnsTopicPolicyResourceType, "ID", mock.MatchedBy(func(input map[string]interface{}) bool {
					return input["id"] == "ID"
				})).Once().Return(&resource.Resource{
					Id:   "ID",
					Type: awsresource.AwsSnsTopicPolicyResourceType,
					Attrs: &resource.Attributes{
						"arn":    "arn",
						"id":     "ID",
						"policy": "{\"policy\":\"coucou\"}",
					},
				}, nil)
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			factory := &dctlresource.MockResourceFactory{}
			if tt.mock != nil {
				tt.mock(factory)
			}

			repo := testresource.InitFakeSchemaRepository("aws", "3.19.0")
			awsresource.InitResourcesMetadata(repo)

			m := NewAwsSNSTopicPolicyExpander(factory, repo)
			if err := m.Execute(&[]*resource.Resource{}, tt.resourcesFromState); (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
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
