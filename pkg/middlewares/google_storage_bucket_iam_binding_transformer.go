package middlewares

import (
	"fmt"

	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/google"
)

// GoogleStorageBucketIAMBindingTransformer Transforms Bucket IAM binding in bucket iam member to ease comparison.
type GoogleStorageBucketIAMBindingTransformer struct {
	resourceFactory resource.ResourceFactory
}

func NewGoogleStorageBucketIAMBindingTransformer(resourceFactory resource.ResourceFactory) *GoogleStorageBucketIAMBindingTransformer {
	return &GoogleStorageBucketIAMBindingTransformer{resourceFactory}
}

func (m *GoogleStorageBucketIAMBindingTransformer) Execute(_, resourcesFromState *[]*resource.Resource) error {

	resources := make([]*resource.Resource, 0)

	for _, stateRes := range *resourcesFromState {
		// Ignore all resources other than BucketIamBinding
		if stateRes.ResourceType() != google.GoogleStorageBucketIamBindingResourceType {
			resources = append(resources, stateRes)
			continue
		}

		bucket := *stateRes.Attrs.GetString("bucket")
		roleName := *stateRes.Attrs.GetString("role")
		members, _ := stateRes.Attrs.Get("members")

		for _, member := range members.([]interface{}) {
			id := fmt.Sprintf("%s/%s/%s", bucket, roleName, member)
			resources = append(
				resources,
				m.resourceFactory.CreateAbstractResource(
					google.GoogleStorageBucketIamMemberResourceType,
					id,
					map[string]interface{}{
						"id":     id,
						"bucket": bucket,
						"role":   roleName,
						"member": member.(string),
					},
				),
			)
		}
	}

	*resourcesFromState = resources

	return nil
}
