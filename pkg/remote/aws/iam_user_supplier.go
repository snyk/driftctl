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

type IamUserSupplier struct {
	reader       terraform.ResourceReader
	deserializer *resource.Deserializer
	repo         repository.IAMRepository
	runner       *terraform.ParallelResourceReader
}

func NewIamUserSupplier(provider *AWSTerraformProvider, deserializer *resource.Deserializer, repo repository.IAMRepository) *IamUserSupplier {
	return &IamUserSupplier{
		provider,
		deserializer,
		repo,
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s *IamUserSupplier) SuppliedType() resource.ResourceType {
	return resourceaws.AwsIamUserResourceType
}

func (s *IamUserSupplier) Resources() ([]resource.Resource, error) {
	users, err := s.repo.ListAllUsers()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, s.SuppliedType())
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
	return s.deserializer.Deserialize(s.SuppliedType(), results)
}

func (s *IamUserSupplier) readUser(user *iam.User) (cty.Value, error) {
	res, err := s.reader.ReadResource(
		terraform.ReadResourceArgs{
			Ty: s.SuppliedType(),
			ID: *user.UserName,
		},
	)
	if err != nil {
		logrus.Warnf("Error reading iam user %s[%s]: %+v", *user.UserName, s.SuppliedType(), err)
		return cty.NilVal, err
	}

	return *res, nil
}
