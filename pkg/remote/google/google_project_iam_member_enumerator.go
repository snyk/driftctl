package google

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/snyk/driftctl/pkg/remote/google/repository"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/google"
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
	errorsByProject := make(map[string]error)

	bindingsByProject, errorsByProject := e.repository.ListProjectsBindings()

	for project, bindings := range bindingsByProject {
		if val, ok := errorsByProject[project]; ok {
			logrus.WithFields(logrus.Fields{
				"project": project,
				"error":   val.Error(),
			}).Debug("When trying to get project IAM members")
		}
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

	return results, nil
}
