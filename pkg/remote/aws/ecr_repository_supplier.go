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
	client       repository.ECRRepository
	runner       *terraform.ParallelResourceReader
}

func NewECRRepositorySupplier(provider *AWSTerraformProvider, deserializer *resource.Deserializer, repo repository.ECRRepository) *ECRRepositorySupplier {
	return &ECRRepositorySupplier{
		provider,
		deserializer,
		repo,
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (r *ECRRepositorySupplier) SuppliedType() resource.ResourceType {
	return aws.AwsEcrRepositoryResourceType
}

func (r *ECRRepositorySupplier) Resources() ([]resource.Resource, error) {
	repositories, err := r.client.ListAllRepositories()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, r.SuppliedType())
	}

	for _, repository := range repositories {
		repository := repository
		r.runner.Run(func() (cty.Value, error) {
			return r.readRepository(repository)
		})
	}

	retrieve, err := r.runner.Wait()
	if err != nil {
		return nil, err
	}

	return r.deserializer.Deserialize(r.SuppliedType(), retrieve)
}

func (r *ECRRepositorySupplier) readRepository(repository *ecr.Repository) (cty.Value, error) {
	val, err := r.reader.ReadResource(terraform.ReadResourceArgs{
		ID: *repository.RepositoryName,
		Ty: r.SuppliedType(),
	})
	if err != nil {
		logrus.Error(err)
		return cty.NilVal, err
	}
	return *val, nil
}
