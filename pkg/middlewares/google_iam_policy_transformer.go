package middlewares

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/google"
)

// GoogleStorageBucketIAMPolicyTransformer Transforms Bucket IAM policy in bucket iam binding to ease comparison.
type GoogleStorageBucketIAMPolicyTransformer struct {
	resourceFactory resource.ResourceFactory
	resFieldByType  map[string]string // map of the field to add to resource attribute for all supported type
}

func NewGoogleIAMPolicyTransformer(resourceFactory resource.ResourceFactory) *GoogleStorageBucketIAMPolicyTransformer {
	return &GoogleStorageBucketIAMPolicyTransformer{
		resourceFactory,
		map[string]string{
			google.GoogleStorageBucketIamPolicyResourceType: "bucket",
			google.GoogleProjectIamPolicyResourceType:       "project",
		}}
}

func (m *GoogleStorageBucketIAMPolicyTransformer) Execute(_, resourcesFromState *[]*resource.Resource) error {

	resources := make([]*resource.Resource, 0)

	for _, stateRes := range *resourcesFromState {
		// Ignore all resources with type not in resFieldByType map
		resType := stateRes.ResourceType()
		resField, supported := m.resFieldByType[resType]
		if !supported {
			resources = append(resources, stateRes)
			continue
		}

		resName := *stateRes.Attrs.GetString(resField)
		policyJSON := *stateRes.Attrs.GetString("policy_data")

		policies := policyDataType{}
		err := json.Unmarshal([]byte(policyJSON), &policies)
		if err != nil {
			return err
		}

		for _, policy := range policies.Bindings {
			roleName := policy["role"].(string)
			members := policy["members"].([]interface{})
			for _, member := range members {
				id := fmt.Sprintf("%s/%s/%s", resName, roleName, member)
				resources = append(
					resources,
					m.resourceFactory.CreateAbstractResource(
						fmt.Sprintf("%s_member", strings.TrimSuffix(resType, "_policy")),
						id,
						map[string]interface{}{
							"id":     id,
							resField: resName,
							"role":   roleName,
							"member": member.(string),
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
	Bindings []map[string]interface{}
}
