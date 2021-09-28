package middlewares

import (
	"fmt"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/google"
)

// GoogleStorageBucketIAMMemberTransformer Transforms Bucket IAM Member in bucket iam binding to ease comparison.
type GoogleStorageBucketIAMMemberTransformer struct {
	resourceFactory resource.ResourceFactory
}

func NewGoogleStorageBucketIAMMemberTransformer(resourceFactory resource.ResourceFactory) *GoogleStorageBucketIAMMemberTransformer {
	return &GoogleStorageBucketIAMMemberTransformer{resourceFactory}
}

func (m *GoogleStorageBucketIAMMemberTransformer) Execute(_, resourcesFromState *[]*resource.Resource) error {

	resources := make([]*resource.Resource, 0)
	bindings := map[string]map[string][]string{} // map[BUCKETID][ROLENAME][]MEMBERS

	for _, stateRes := range *resourcesFromState {
		// Ignore all resources other than BucketIamBinding
		if stateRes.ResourceType() != google.GoogleStorageBucketIamMemberResourceType {
			resources = append(resources, stateRes)
			continue
		}

		bucket := *stateRes.Attrs.GetString("bucket")
		roleName := *stateRes.Attrs.GetString("role")
		member := *stateRes.Attrs.GetString("member")

		bucketBindings := bindings[bucket]
		if bucketBindings == nil {
			bucketBindings = make(map[string][]string)
		}

		roleBindings := bucketBindings[roleName]
		if roleBindings == nil {
			roleBindings = make([]string, 0, 1)
		}

		roleBindings = append(roleBindings, member)

		bucketBindings[roleName] = roleBindings
		bindings[bucket] = bucketBindings
	}

	for bucket, roleBindings := range bindings {
		for roleName, members := range roleBindings {
			id := fmt.Sprintf("%s/%s", bucket, roleName)
			resources = append(
				resources,
				m.resourceFactory.CreateAbstractResource(
					google.GoogleStorageBucketIamBindingResourceType,
					id,
					map[string]interface{}{
						"id":      id,
						"bucket":  bucket,
						"role":    roleName,
						"members": members,
					},
				),
			)
		}

	}

	*resourcesFromState = resources

	return nil
}
