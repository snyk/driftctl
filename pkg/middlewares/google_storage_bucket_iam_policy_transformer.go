package middlewares

import (
	"encoding/json"
	"fmt"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/google"
)

// GoogleStorageBucketIAMPolicyTransformer Transforms Bucket IAM policy in bucket iam binding to ease comparison.
type GoogleStorageBucketIAMPolicyTransformer struct {
	resourceFactory resource.ResourceFactory
}

func NewGoogleStorageBucketIAMPolicyTransformer(resourceFactory resource.ResourceFactory) *GoogleStorageBucketIAMPolicyTransformer {
	return &GoogleStorageBucketIAMPolicyTransformer{resourceFactory}
}

func (m *GoogleStorageBucketIAMPolicyTransformer) Execute(_, resourcesFromState *[]*resource.Resource) error {

	resources := make([]*resource.Resource, 0)

	for _, stateRes := range *resourcesFromState {
		// Ignore all resources other than BucketIamBinding
		if stateRes.ResourceType() != google.GoogleStorageBucketIamPolicyResourceType {
			resources = append(resources, stateRes)
			continue
		}

		bucket := *stateRes.Attrs.GetString("bucket")
		policyJSON := *stateRes.Attrs.GetString("policy_data")

		policies := policyDataType{}
		err := json.Unmarshal([]byte(policyJSON), &policies)
		if err != nil {
			return err
		}

		for _, policy := range policies.Bindings {
			roleName := policy.Role
			members := policy.Members
			for _, member := range members {
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
							"member": member,
						},
					),
				)
			}
		}
	}

	*resourcesFromState = resources

	return nil
}

type policyDataType struct {
	Bindings []struct {
		Members []string
		Role    string
	}
}
