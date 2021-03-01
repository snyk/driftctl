package github

import (
	"fmt"

	"github.com/cloudskiff/driftctl/pkg/remote/deserializer"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourcegithub "github.com/cloudskiff/driftctl/pkg/resource/github"
	ghdeserializer "github.com/cloudskiff/driftctl/pkg/resource/github/deserializer"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

type GithubTeamSupplier struct {
	reader       terraform.ResourceReader
	deserializer deserializer.CTYDeserializer
	repository   GithubRepository
	runner       *terraform.ParallelResourceReader
}

func NewGithubTeamSupplier(provider *GithubTerraformProvider, repository GithubRepository) *GithubTeamSupplier {
	return &GithubTeamSupplier{
		provider,
		ghdeserializer.NewGithubTeamDeserializer(),
		repository,
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s GithubTeamSupplier) Resources() ([]resource.Resource, error) {

	resourceList, err := s.repository.ListTeams()

	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, resourcegithub.GithubTeamResourceType)
	}

	for _, team := range resourceList {
		team := team
		s.runner.Run(func() (cty.Value, error) {
			completeResource, err := s.reader.ReadResource(terraform.ReadResourceArgs{
				Ty: resourcegithub.GithubTeamResourceType,
				ID: fmt.Sprintf("%d", team.DatabaseId),
			})
			if err != nil {
				logrus.Warnf("Error reading %d[%s]: %+v", team.DatabaseId, resourcegithub.GithubTeamResourceType, err)
				return cty.NilVal, err
			}
			return *completeResource, nil
		})
	}

	results, err := s.runner.Wait()
	if err != nil {
		return nil, err
	}

	return s.deserializer.Deserialize(results)
}
