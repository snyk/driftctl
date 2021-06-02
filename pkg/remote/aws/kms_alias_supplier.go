package aws

import (
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/cloudskiff/driftctl/pkg/terraform"
)

type KMSAliasSupplier struct {
	reader       terraform.ResourceReader
	deserializer *resource.Deserializer
	client       repository.KMSRepository
	runner       *terraform.ParallelResourceReader
}

func NewKMSAliasSupplier(provider *AWSTerraformProvider, deserializer *resource.Deserializer) *KMSAliasSupplier {
	return &KMSAliasSupplier{
		provider,
		deserializer,
		repository.NewKMSRepository(provider.session),
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s *KMSAliasSupplier) Resources() ([]resource.Resource, error) {
	aliases, err := s.client.ListAllAliases()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, aws.AwsKmsAliasResourceType)
	}

	for _, alias := range aliases {
		alias := alias
		s.runner.Run(func() (cty.Value, error) {
			return s.readAlias(alias)
		})
	}

	retrieve, err := s.runner.Wait()
	if err != nil {
		return nil, err
	}

	return s.deserializer.Deserialize(aws.AwsKmsAliasResourceType, retrieve)
}

func (s *KMSAliasSupplier) readAlias(alias *kms.AliasListEntry) (cty.Value, error) {
	val, err := s.reader.ReadResource(terraform.ReadResourceArgs{
		ID: *alias.AliasName,
		Ty: aws.AwsKmsAliasResourceType,
	})
	if err != nil {
		logrus.Error(err)
		return cty.NilVal, err
	}
	return *val, nil
}
