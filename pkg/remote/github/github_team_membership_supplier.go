package github

import (
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourcegithub "github.com/cloudskiff/driftctl/pkg/resource/github"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

type GithubTeamMembershipSupplier struct {
	reader       terraform.ResourceReader
	deserializer *resource.Deserializer
	repository   GithubRepository
	runner       *terraform.ParallelResourceReader
}

func NewGithubTeamMembershipSupplier(provider *GithubTerraformProvider, repository GithubRepository, deserializer *resource.Deserializer) *GithubTeamMembershipSupplier {
	return &GithubTeamMembershipSupplier{
		provider,
		deserializer,
		repository,
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s *GithubTeamMembershipSupplier) SuppliedType() resource.ResourceType {
	return resourcegithub.GithubTeamMembershipResourceType
}

func (s GithubTeamMembershipSupplier) Resources() ([]resource.Resource, error) {
	resourceList, err := s.repository.ListTeamMemberships()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, s.SuppliedType())
	}

	for _, id := range resourceList {
		id := id
		s.runner.Run(func() (cty.Value, error) {
			return s.readTeamMembership(id)
		})
	}

	results, err := s.runner.Wait()
	if err != nil {
		return nil, err
	}

	return s.deserializer.Deserialize(s.SuppliedType(), results)
}

func (s GithubTeamMembershipSupplier) readTeamMembership(id string) (cty.Value, error) {
	val, err := s.reader.ReadResource(terraform.ReadResourceArgs{
		ID: id,
		Ty: s.SuppliedType(),
	})
	if err != nil {
		logrus.Warnf("Error reading %s[%s]: %+v", id, s.SuppliedType(), err)
		return cty.NilVal, err
	}
	return *val, nil
}
