package github

import (
	"github.com/cloudskiff/driftctl/pkg/remote/deserializer"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourcegithub "github.com/cloudskiff/driftctl/pkg/resource/github"
	ghdeserializer "github.com/cloudskiff/driftctl/pkg/resource/github/deserializer"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

type GithubBranchProtectionSupplier struct {
	reader       terraform.ResourceReader
	deserializer deserializer.CTYDeserializer
	repository   GithubRepository
	runner       *terraform.ParallelResourceReader
}

func NewGithubBranchProtectionSupplier(provider *GithubTerraformProvider, repository GithubRepository) *GithubBranchProtectionSupplier {
	return &GithubBranchProtectionSupplier{
		provider,
		ghdeserializer.NewGithubBranchProtectionDeserializer(),
		repository,
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s GithubBranchProtectionSupplier) Resources() ([]resource.Resource, error) {

	resourceList, err := s.repository.ListBranchProtection()

	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, resourcegithub.GithubBranchProtectionResourceType)
	}

	for _, id := range resourceList {
		id := id
		s.runner.Run(func() (cty.Value, error) {
			completeResource, err := s.reader.ReadResource(terraform.ReadResourceArgs{
				Ty: resourcegithub.GithubBranchProtectionResourceType,
				ID: id,
			})
			if err != nil {
				logrus.Warnf("Error reading %s[%s]: %+v", id, resourcegithub.GithubBranchProtectionResourceType, err)
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
