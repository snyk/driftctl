package middlewares

import (
	"fmt"
	"strings"

	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/google"
)

// GoogleIAMBindingTransformer Transforms Bucket IAM binding in bucket iam member to ease comparison.
type GoogleIAMBindingTransformer struct {
	resourceFactory resource.ResourceFactory
	resFieldByType  map[string]string // map of the field to add to resource attribute for all supported type
}

func NewGoogleIAMBindingTransformer(resourceFactory resource.ResourceFactory) *GoogleIAMBindingTransformer {
	return &GoogleIAMBindingTransformer{
		resourceFactory,
		map[string]string{
			google.GoogleStorageBucketIamBindingResourceType: "bucket",
			google.GoogleProjectIamBindingResourceType:       "project",
		},
	}
}

func (m *GoogleIAMBindingTransformer) Execute(_, resourcesFromState *[]*resource.Resource) error {

	resources := make([]*resource.Resource, 0)

	for _, stateRes := range *resourcesFromState {
		// Ignore all resources other than IamBinding
		resType := stateRes.ResourceType()
		resField, supported := m.resFieldByType[resType]
		if !supported {
			resources = append(resources, stateRes)
			continue
		}

		resName := *stateRes.Attrs.GetString(resField)
		roleName := *stateRes.Attrs.GetString("role")
		members, exist := stateRes.Attrs.Get("members")

		if !exist || members == nil {
			continue
		}

		for _, member := range members.([]interface{}) {
			id := fmt.Sprintf("%s/%s/%s", resName, roleName, member)
			resources = append(
				resources,
				m.resourceFactory.CreateAbstractResource(
					fmt.Sprintf("%s_member", strings.TrimSuffix(resType, "_binding")),
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

	*resourcesFromState = resources

	return nil
}
