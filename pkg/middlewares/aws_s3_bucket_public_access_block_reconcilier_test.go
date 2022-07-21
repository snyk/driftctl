package middlewares

import (
	"testing"

	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
	"github.com/stretchr/testify/assert"
)

func TestAwsS3BucketPublicAccessBlockReconciler(t *testing.T) {
	tests := []struct {
		name               string
		remoteResources    []*resource.Resource
		resourcesFromState []*resource.Resource
		expected           []*resource.Resource
		wantErr            assert.ErrorAssertionFunc
	}{
		{
			name: "ensure we ignore resources that are not of the good type",
			remoteResources: []*resource.Resource{
				{
					Id:   "should_not_be_skipped_because_wrong_type",
					Type: "wrong_type",
					Attrs: &resource.Attributes{
						"block_public_acls":       false,
						"block_public_policy":     false,
						"ignore_public_acls":      false,
						"restrict_public_buckets": false,
					},
				},
			},
			resourcesFromState: []*resource.Resource{
				{
					Id:   "should_not_be_skipped_because_wrong_type",
					Type: "wrong_type",
					Attrs: &resource.Attributes{
						"block_public_acls":       false,
						"block_public_policy":     false,
						"ignore_public_acls":      false,
						"restrict_public_buckets": false,
					},
				},
			},
			expected: []*resource.Resource{
				{
					Id:   "should_not_be_skipped_because_wrong_type",
					Type: "wrong_type",
					Attrs: &resource.Attributes{
						"block_public_acls":       false,
						"block_public_policy":     false,
						"ignore_public_acls":      false,
						"restrict_public_buckets": false,
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "ensure we do not skip non default ones",
			remoteResources: []*resource.Resource{
				{
					Id:   "should_be_present_because_non_default",
					Type: aws.AwsS3BucketPublicAccessBlockResourceType,
					Attrs: &resource.Attributes{
						"block_public_acls":       true,
						"block_public_policy":     false,
						"ignore_public_acls":      false,
						"restrict_public_buckets": false,
					},
				},
			},
			resourcesFromState: []*resource.Resource{
				{
					Id:   "should_be_present_because_non_default",
					Type: aws.AwsS3BucketPublicAccessBlockResourceType,
					Attrs: &resource.Attributes{
						"block_public_acls":       true,
						"block_public_policy":     false,
						"ignore_public_acls":      false,
						"restrict_public_buckets": false,
					},
				},
			},
			expected: []*resource.Resource{
				{
					Id:   "should_be_present_because_non_default",
					Type: aws.AwsS3BucketPublicAccessBlockResourceType,
					Attrs: &resource.Attributes{
						"block_public_acls":       true,
						"block_public_policy":     false,
						"ignore_public_acls":      false,
						"restrict_public_buckets": false,
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "ensure default public access block are removed",
			remoteResources: []*resource.Resource{
				{
					Id:   "should_be_skipped_because_default",
					Type: aws.AwsS3BucketPublicAccessBlockResourceType,
					Attrs: &resource.Attributes{
						"block_public_acls":       false,
						"block_public_policy":     false,
						"ignore_public_acls":      false,
						"restrict_public_buckets": false,
					},
				},
				{
					Id:   "should_be_skipped_because_nil_values",
					Type: aws.AwsS3BucketPublicAccessBlockResourceType,
					Attrs: &resource.Attributes{
						"block_public_acls": false,
					},
				},
				{
					Id:   "should_not_be_skipped_because_exist_in_iac",
					Type: aws.AwsS3BucketPublicAccessBlockResourceType,
					Attrs: &resource.Attributes{
						"block_public_acls": false,
					},
				},
			},
			resourcesFromState: []*resource.Resource{
				{
					Id:   "should_not_be_skipped_because_exist_in_iac",
					Type: aws.AwsS3BucketPublicAccessBlockResourceType,
					Attrs: &resource.Attributes{
						"block_public_acls": false,
					},
				},
			},
			expected: []*resource.Resource{
				{
					Id:   "should_not_be_skipped_because_exist_in_iac",
					Type: aws.AwsS3BucketPublicAccessBlockResourceType,
					Attrs: &resource.Attributes{
						"block_public_acls": false,
					},
				},
			},
			wantErr: nil,
		},
	}

	r := NewAwsS3BucketPublicAccessBlockReconciler()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = r.Execute(&tt.remoteResources, &tt.resourcesFromState)
			assert.Equal(t, tt.remoteResources, tt.expected)
			assert.Equal(t, tt.resourcesFromState, tt.expected)
		})
	}
}
