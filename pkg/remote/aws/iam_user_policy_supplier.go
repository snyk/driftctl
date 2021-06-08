package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/cloudskiff/driftctl/pkg/terraform"

	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

type IamUserPolicySupplier struct {
	reader       terraform.ResourceReader
	deserializer *resource.Deserializer
	repo         repository.IAMRepository
	runner       *terraform.ParallelResourceReader
}

func NewIamUserPolicySupplier(provider *AWSTerraformProvider, deserializer *resource.Deserializer, repo repository.IAMRepository) *IamUserPolicySupplier {
	return &IamUserPolicySupplier{
		provider,
		deserializer,
		repo,
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s *IamUserPolicySupplier) SuppliedType() resource.ResourceType {
	return resourceaws.AwsIamUserPolicyResourceType
}

func (s *IamUserPolicySupplier) Resources() ([]resource.Resource, error) {
	users, err := s.repo.ListAllUsers()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationErrorWithType(err, s.SuppliedType(), resourceaws.AwsIamUserResourceType)
	}
	policies, err := s.repo.ListAllUserPolicies(users)
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, s.SuppliedType())
	}
	results := make([]cty.Value, 0)
	if len(policies) > 0 {
		for _, policy := range policies {
			p := policy
			s.runner.Run(func() (cty.Value, error) {
				return s.readUserPolicy(p)
			})
		}
		results, err = s.runner.Wait()
		if err != nil {
			return nil, err
		}
	}
	return s.deserializer.Deserialize(s.SuppliedType(), results)
}

func (s *IamUserPolicySupplier) readUserPolicy(policyName string) (cty.Value, error) {
	res, err := s.reader.ReadResource(
		terraform.ReadResourceArgs{
			Ty: s.SuppliedType(),
			ID: policyName,
		},
	)
	if err != nil {
		logrus.Warnf("Error reading iam user policy %s[%s]: %+v", policyName, resourceaws.AwsIamUserResourceType, err)
		return cty.NilVal, err
	}

	return *res, nil
}
