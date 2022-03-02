package repository

import (
	"errors"
	"fmt"
	"strings"

	"github.com/snyk/driftctl/pkg/remote/cache"
	"github.com/snyk/driftctl/pkg/remote/google/config"
	"google.golang.org/api/cloudresourcemanager/v1"
)

type CloudResourceManagerRepository interface {
	ListProjectsBindings() (map[string]map[string][]string, error)
}

type cloudResourceManagerRepository struct {
	service *cloudresourcemanager.Service
	config  config.GCPTerraformConfig
	cache   cache.Cache
}

func NewCloudResourceManagerRepository(service *cloudresourcemanager.Service, config config.GCPTerraformConfig, cache cache.Cache) CloudResourceManagerRepository {
	return &cloudResourceManagerRepository{
		service: service,
		config:  config,
		cache:   cache,
	}
}

func (s *cloudResourceManagerRepository) ListProjectsBindings() (map[string]map[string][]string, error) {

	bindingsByProject := make(map[string]map[string][]string)
	errorsByProject := make(map[string]error)
	var erorString string

	for _, scope := range s.config.Scopes {
		if strings.Contains(scope, "projects/") {
			project := strings.Split(scope, "projects/")[1]
			request := new(cloudresourcemanager.GetIamPolicyRequest)
			policy, err := s.service.Projects.GetIamPolicy(project, request).Do()
			if err != nil {
				errorsByProject[project] = err
				bindingsByProject[project] = nil
				continue
			}

			bindings := make(map[string][]string)
			for _, binding := range policy.Bindings {
				bindings[binding.Role] = binding.Members
			}

			bindingsByProject[project] = bindings

			s.cache.Put("ListProjectsBindings", bindingsByProject)

		}
	}

	if len(errorsByProject) > 0 {
		for project, errval := range errorsByProject {
			erorString = erorString + fmt.Sprintf("Project: %s had the following error: %s; ", project, errval.Error())
		}
		return bindingsByProject, errors.New(erorString)
	}

	return bindingsByProject, nil
}
