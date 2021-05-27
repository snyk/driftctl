package aws

import (
	"github.com/aws/aws-sdk-go/service/iam"

	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	"github.com/cloudskiff/driftctl/pkg/remote/deserializer"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"
	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"
	"github.com/cloudskiff/driftctl/pkg/terraform"

	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

type IamUserSupplier struct {
	reader       terraform.ResourceReader
	deserializer deserializer.CTYDeserializer
	repo         repository.IAMRepository
	runner       *terraform.ParallelResourceReader
}

func NewIamUserSupplier(provider *AWSTerraformProvider) *IamUserSupplier {
	return &IamUserSupplier{
		provider,
		awsdeserializer.NewIamUserDeserializer(),
		repository.NewIAMRepository(provider.session),
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s *IamUserSupplier) Resources() ([]resource.Resource, error) {
	users, err := s.repo.ListAllUsers()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, resourceaws.AwsIamUserResourceType)
	}
	results := make([]cty.Value, 0)
	if len(users) > 0 {
		for _, user := range users {
			u := *user
			s.runner.Run(func() (cty.Value, error) {
				return s.readUser(&u)
			})
		}
		results, err = s.runner.Wait()
		if err != nil {
			return nil, err
		}
	}
	return s.deserializer.Deserialize(results)
}

func (s *IamUserSupplier) readUser(user *iam.User) (cty.Value, error) {
	res, err := s.reader.ReadResource(
		terraform.ReadResourceArgs{
			Ty: resourceaws.AwsIamUserResourceType,
			ID: *user.UserName,
		},
	)
	if err != nil {
		logrus.Warnf("Error reading iam user %s[%s]: %+v", *user.UserName, resourceaws.AwsIamUserResourceType, err)
		return cty.NilVal, err
	}

	return *res, nil
}
