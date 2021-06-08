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

type IamRolePolicySupplier struct {
	reader       terraform.ResourceReader
	deserializer *resource.Deserializer
	repo         repository.IAMRepository
	runner       *terraform.ParallelResourceReader
}

func NewIamRolePolicySupplier(provider *AWSTerraformProvider, deserializer *resource.Deserializer, repo repository.IAMRepository) *IamRolePolicySupplier {
	return &IamRolePolicySupplier{
		provider,
		deserializer,
		repo,
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s *IamRolePolicySupplier) SuppliedType() resource.ResourceType {
	return resourceaws.AwsIamRolePolicyResourceType
}

func (s *IamRolePolicySupplier) Resources() ([]resource.Resource, error) {
	roles, err := s.repo.ListAllRoles()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationErrorWithType(err, s.SuppliedType(), resourceaws.AwsIamRoleResourceType)
	}
	policies, err := s.repo.ListAllRolePolicies(roles)
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, s.SuppliedType())
	}
	results := make([]cty.Value, 0)
	if len(policies) > 0 {
		for _, policy := range policies {
			p := policy
			s.runner.Run(func() (cty.Value, error) {
				return s.readRolePolicy(p)
			})
		}
		results, err = s.runner.Wait()
		if err != nil {
			return nil, err
		}
	}

	return s.deserializer.Deserialize(s.SuppliedType(), results)
}

func (s *IamRolePolicySupplier) readRolePolicy(policyName string) (cty.Value, error) {
	res, err := s.reader.ReadResource(
		terraform.ReadResourceArgs{
			Ty: s.SuppliedType(),
			ID: policyName,
		},
	)
	if err != nil {
		logrus.Warnf("Error reading iam role policy %s[%s]: %+v", policyName, s.SuppliedType(), err)
		return cty.NilVal, err
	}

	return *res, nil
}
