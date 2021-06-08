package github

import (
	"fmt"

	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourcegithub "github.com/cloudskiff/driftctl/pkg/resource/github"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

type GithubTeamSupplier struct {
	reader       terraform.ResourceReader
	deserializer *resource.Deserializer
	repository   GithubRepository
	runner       *terraform.ParallelResourceReader
}

func NewGithubTeamSupplier(provider *GithubTerraformProvider, repository GithubRepository, deserializer *resource.Deserializer) *GithubTeamSupplier {
	return &GithubTeamSupplier{
		provider,
		deserializer,
		repository,
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s *GithubTeamSupplier) SuppliedType() resource.ResourceType {
	return resourcegithub.GithubTeamResourceType
}

func (s *GithubTeamSupplier) Resources() ([]resource.Resource, error) {

	resourceList, err := s.repository.ListTeams()

	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, s.SuppliedType())
	}

	for _, team := range resourceList {
		team := team
		s.runner.Run(func() (cty.Value, error) {
			completeResource, err := s.reader.ReadResource(terraform.ReadResourceArgs{
				Ty: s.SuppliedType(),
				ID: fmt.Sprintf("%d", team.DatabaseId),
			})
			if err != nil {
				logrus.Warnf("Error reading %d[%s]: %+v", team.DatabaseId, s.SuppliedType(), err)
				return cty.NilVal, err
			}
			return *completeResource, nil
		})
	}

	results, err := s.runner.Wait()
	if err != nil {
		return nil, err
	}

	return s.deserializer.Deserialize(s.SuppliedType(), results)
}
