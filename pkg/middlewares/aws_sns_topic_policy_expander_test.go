package middlewares

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"

	awsresource "github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	testresource "github.com/cloudskiff/driftctl/test/resource"

	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/r3labs/diff/v2"

	"github.com/cloudskiff/driftctl/pkg/resource"
)

func TestAwsSNSTopicPolicyExpander_Execute(t *testing.T) {
	tests := []struct {
		name               string
		resourcesFromState *[]resource.Resource
		expected           *[]resource.Resource
		mock               func(factory *terraform.MockResourceFactory)
		wantErr            bool
	}{
		{
			name: "Inline policy no attached policy",
			resourcesFromState: &[]resource.Resource{
				&resource.AbstractResource{
					Id:   "ID",
					Type: awsresource.AwsSnsTopicResourceType,
					Attrs: &resource.Attributes{
						"arn":    "arn",
						"id":     "ID",
						"policy": "{\"policy\":\"coucou\"}",
					},
				},
			},
			expected: &[]resource.Resource{
				&resource.AbstractResource{
					Id:   "ID",
					Type: awsresource.AwsSnsTopicResourceType,
					Attrs: &resource.Attributes{
						"arn": "arn",
						"id":  "ID",
					},
				},
				&resource.AbstractResource{
					Id:   "ID",
					Type: awsresource.AwsSnsTopicPolicyResourceType,
					Attrs: &resource.Attributes{
						"arn":    "arn",
						"id":     "ID",
						"policy": "{\"policy\":\"coucou\"}",
					},
				},
			},
			mock: func(factory *terraform.MockResourceFactory) {
				factory.On("CreateAbstractResource", "ID", awsresource.AwsSnsTopicPolicyResourceType, mock.MatchedBy(func(input map[string]interface{}) bool {
					return input["id"] == "ID"
				})).Once().Return(&resource.AbstractResource{
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
			resourcesFromState: &[]resource.Resource{
				&resource.AbstractResource{
					Id:   "ID",
					Type: awsresource.AwsSnsTopicResourceType,
					Attrs: &resource.Attributes{
						"arn": "arn",
						"id":  "ID",
					},
				},
				&resource.AbstractResource{
					Id:   "ID",
					Type: awsresource.AwsSnsTopicPolicyResourceType,
					Attrs: &resource.Attributes{
						"arn":    "arn",
						"id":     "ID",
						"policy": "{\"policy\":\"coucou\"}",
					},
				},
			},
			expected: &[]resource.Resource{
				&resource.AbstractResource{
					Id:   "ID",
					Type: awsresource.AwsSnsTopicResourceType,
					Attrs: &resource.Attributes{
						"arn": "arn",
						"id":  "ID",
					},
				},
				&resource.AbstractResource{
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
			resourcesFromState: &[]resource.Resource{
				&resource.AbstractResource{
					Id:   "ID",
					Type: awsresource.AwsSnsTopicResourceType,
					Attrs: &resource.Attributes{
						"arn":    "arn",
						"id":     "ID",
						"policy": "{\"policy\":\"coucou\"}",
					},
				},
				&resource.AbstractResource{
					Id:   "ID",
					Type: awsresource.AwsSnsTopicPolicyResourceType,
					Attrs: &resource.Attributes{
						"arn":    "arn",
						"id":     "ID",
						"policy": "{\"policy\":\"coucou\"}",
					},
				},
			},
			expected: &[]resource.Resource{
				&resource.AbstractResource{
					Id:   "ID",
					Type: awsresource.AwsSnsTopicResourceType,
					Attrs: &resource.Attributes{
						"arn": "arn",
						"id":  "ID",
					},
				},
				&resource.AbstractResource{
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
			resourcesFromState: &[]resource.Resource{
				&resource.AbstractResource{
					Id:   "ID",
					Type: awsresource.AwsSnsTopicResourceType,
					Attrs: &resource.Attributes{
						"arn":    "arn",
						"id":     "ID",
						"policy": "{\"policy\":\"coucou\"}",
					},
				},
				&resource.AbstractResource{
					Id:   "ID2",
					Type: awsresource.AwsSnsTopicPolicyResourceType,
					Attrs: &resource.Attributes{
						"arn":    "arn2",
						"id":     "ID2",
						"policy": "{\"policy\":\"coucou2\"}",
					},
				},
			},
			expected: &[]resource.Resource{
				&resource.AbstractResource{
					Id:   "ID",
					Type: awsresource.AwsSnsTopicResourceType,
					Attrs: &resource.Attributes{
						"arn": "arn",
						"id":  "ID",
					},
				},
				&resource.AbstractResource{
					Id:   "ID",
					Type: awsresource.AwsSnsTopicPolicyResourceType,
					Attrs: &resource.Attributes{
						"arn":    "arn",
						"id":     "ID",
						"policy": "{\"policy\":\"coucou\"}",
					},
				},
				&resource.AbstractResource{
					Id:   "ID2",
					Type: awsresource.AwsSnsTopicPolicyResourceType,
					Attrs: &resource.Attributes{
						"arn":    "arn2",
						"id":     "ID2",
						"policy": "{\"policy\":\"coucou2\"}",
					},
				},
			},
			mock: func(factory *terraform.MockResourceFactory) {
				factory.On("CreateAbstractResource", "ID", awsresource.AwsSnsTopicPolicyResourceType, mock.MatchedBy(func(input map[string]interface{}) bool {
					return input["id"] == "ID"
				})).Once().Return(&resource.AbstractResource{
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

			factory := &terraform.MockResourceFactory{}
			if tt.mock != nil {
				tt.mock(factory)
			}

			repo := testresource.InitFakeSchemaRepository("aws", "3.19.0")
			awsresource.InitResourcesMetadata(repo)

			m := NewAwsSNSTopicPolicyExpander(factory, repo)
			if err := m.Execute(&[]resource.Resource{}, tt.resourcesFromState); (err != nil) != tt.wantErr {
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
