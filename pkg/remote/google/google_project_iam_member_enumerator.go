package google

import (
	"fmt"

	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/remote/google/repository"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/google"
)

type GoogleProjectIamMemberEnumerator struct {
	repository repository.CloudResourceManagerRepository
	factory    resource.ResourceFactory
}

func NewGoogleProjectIamMemberEnumerator(repo repository.CloudResourceManagerRepository, factory resource.ResourceFactory) *GoogleProjectIamMemberEnumerator {
	return &GoogleProjectIamMemberEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *GoogleProjectIamMemberEnumerator) SupportedType() resource.ResourceType {
	return google.GoogleProjectIamMemberResourceType
}

func (e *GoogleProjectIamMemberEnumerator) Enumerate() ([]*resource.Resource, error) {
	results := make([]*resource.Resource, 0)

	bindingsByProject, err := e.repository.ListProjectsBindings()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}
	for project, bindings := range bindingsByProject {
		for roleName, members := range bindings {
			for _, member := range members {
				id := fmt.Sprintf("%s/%s/%s", project, roleName, member)
				results = append(
					results,
					e.factory.CreateAbstractResource(
						string(e.SupportedType()),
						id,
						map[string]interface{}{
							"id":      id,
							"project": project,
							"role":    roleName,
							"member":  member,
						},
					),
				)
			}
		}
	}

	return results, err
}
