package aws

import (
	"github.com/aws/aws-sdk-go/service/iam"

	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/cloudskiff/driftctl/pkg/terraform"

	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

type IamAccessKeySupplier struct {
	reader       terraform.ResourceReader
	deserializer *resource.Deserializer
	repo         repository.IAMRepository
	runner       *terraform.ParallelResourceReader
}

func NewIamAccessKeySupplier(provider *AWSTerraformProvider, deserializer *resource.Deserializer) *IamAccessKeySupplier {
	return &IamAccessKeySupplier{
		provider,
		deserializer,
		repository.NewIAMRepository(provider.session),
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s *IamAccessKeySupplier) Resources() ([]resource.Resource, error) {
	users, err := s.repo.ListAllUsers()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationErrorWithType(err, resourceaws.AwsIamAccessKeyResourceType, resourceaws.AwsIamUserResourceType)
	}
	keys, err := s.repo.ListAllAccessKeys(users)
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, resourceaws.AwsIamAccessKeyResourceType)
	}
	results := make([]cty.Value, 0)
	if len(keys) > 0 {
		for _, key := range keys {
			k := *key
			s.runner.Run(func() (cty.Value, error) {
				return s.readAccessKey(&k)
			})
		}
		results, err = s.runner.Wait()
		if err != nil {
			return nil, err
		}
	}
	return s.deserializer.Deserialize(resourceaws.AwsIamAccessKeyResourceType, results)
}

func (s *IamAccessKeySupplier) readAccessKey(key *iam.AccessKeyMetadata) (cty.Value, error) {
	res, err := s.reader.ReadResource(
		terraform.ReadResourceArgs{
			Ty: resourceaws.AwsIamAccessKeyResourceType,
			ID: *key.AccessKeyId,
			Attributes: map[string]string{
				"user": *key.UserName,
			},
		},
	)
	if err != nil {
		logrus.Warnf("Error reading iam access key %s[%s]: %+v", *key.AccessKeyId, resourceaws.AwsIamAccessKeyResourceType, err)
		return cty.NilVal, err
	}

	return *res, nil
}
