package repository

import (
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
	"github.com/cloudskiff/driftctl/pkg/remote/google/config"
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
	if cachedResults := s.cache.Get("ListProjectsBindings"); cachedResults != nil {
		return cachedResults.(map[string]map[string][]string), nil
	}

	request := new(cloudresourcemanager.GetIamPolicyRequest)
	policy, err := s.service.Projects.GetIamPolicy(s.config.Project, request).Do()
	if err != nil {
		return nil, err
	}

	bindings := make(map[string][]string)

	for _, binding := range policy.Bindings {
		bindings[binding.Role] = binding.Members
	}

	bindingsByProject := make(map[string]map[string][]string)
	bindingsByProject[s.config.Project] = bindings

	s.cache.Put("ListProjectsBindings", bindingsByProject)

	return bindingsByProject, nil
}
