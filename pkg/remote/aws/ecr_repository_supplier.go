package aws

import (
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/cloudskiff/driftctl/pkg/terraform"
)

type ECRRepositorySupplier struct {
	reader       terraform.ResourceReader
	deserializer *resource.Deserializer
	repository   repository.ECRRepository
	runner       *terraform.ParallelResourceReader
}

func NewECRRepositorySupplier(provider *AWSTerraformProvider, deserializer *resource.Deserializer, repository repository.ECRRepository) *ECRRepositorySupplier {
	return &ECRRepositorySupplier{
		provider,
		deserializer,
		repository,
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (r *ECRRepositorySupplier) Resources() ([]resource.Resource, error) {
	repositories, err := r.repository.ListAllRepositories()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, aws.AwsEcrRepositoryResourceType)
	}

	for _, repo := range repositories {
		repo := repo
		r.runner.Run(func() (cty.Value, error) {
			return r.readRepository(repo)
		})
	}

	retrieve, err := r.runner.Wait()
	if err != nil {
		return nil, err
	}

	return r.deserializer.Deserialize(aws.AwsEcrRepositoryResourceType, retrieve)
}

func (r *ECRRepositorySupplier) readRepository(repository *ecr.Repository) (cty.Value, error) {
	val, err := r.reader.ReadResource(terraform.ReadResourceArgs{
		ID: *repository.RepositoryName,
		Ty: aws.AwsEcrRepositoryResourceType,
	})
	if err != nil {
		logrus.Error(err)
		return cty.NilVal, err
	}
	return *val, nil
}
