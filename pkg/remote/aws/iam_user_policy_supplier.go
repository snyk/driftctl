package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"

	"github.com/cloudskiff/driftctl/pkg/remote/deserializer"
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/cloudskiff/driftctl/pkg/terraform"

	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

type IamUserPolicySupplier struct {
	reader       terraform.ResourceReader
	deserializer deserializer.CTYDeserializer
	repo         repository.IAMRepository
	runner       *terraform.ParallelResourceReader
}

func NewIamUserPolicySupplier(provider *AWSTerraformProvider) *IamUserPolicySupplier {
	return &IamUserPolicySupplier{
		provider,
		awsdeserializer.NewIamUserPolicyDeserializer(),
		repository.NewIAMRepository(provider.session),
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s *IamUserPolicySupplier) Resources() ([]resource.Resource, error) {
	users, err := s.repo.ListAllUsers()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationErrorWithType(err, resourceaws.AwsIamUserPolicyResourceType, resourceaws.AwsIamUserResourceType)
	}
	policies, err := s.repo.ListAllUserPolicies(users)
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, resourceaws.AwsIamUserPolicyResourceType)
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
	return s.deserializer.Deserialize(results)
}

func (s *IamUserPolicySupplier) readUserPolicy(policyName string) (cty.Value, error) {
	res, err := s.reader.ReadResource(
		terraform.ReadResourceArgs{
			Ty: resourceaws.AwsIamUserPolicyResourceType,
			ID: policyName,
		},
	)
	if err != nil {
		logrus.Warnf("Error reading iam user policy %s[%s]: %+v", policyName, resourceaws.AwsIamUserResourceType, err)
		return cty.NilVal, err
	}

	return *res, nil
}
