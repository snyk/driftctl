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

type IamPolicySupplier struct {
	reader       terraform.ResourceReader
	deserializer *resource.Deserializer
	repo         repository.IAMRepository
	runner       *terraform.ParallelResourceReader
}

func NewIamPolicySupplier(provider *AWSTerraformProvider, deserializer *resource.Deserializer, repo repository.IAMRepository) *IamPolicySupplier {
	return &IamPolicySupplier{
		provider,
		deserializer,
		repo,
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s *IamPolicySupplier) SuppliedType() resource.ResourceType {
	return resourceaws.AwsIamPolicyResourceType
}

func (s *IamPolicySupplier) Resources() ([]resource.Resource, error) {
	policies, err := s.repo.ListAllPolicies()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, s.SuppliedType())
	}
	results := make([]cty.Value, 0)
	if len(policies) > 0 {
		for _, policy := range policies {
			u := *policy
			s.runner.Run(func() (cty.Value, error) {
				return s.readPolicy(&u)
			})
		}
		results, err = s.runner.Wait()
		if err != nil {
			return nil, err
		}
	}
	return s.deserializer.Deserialize(s.SuppliedType(), results)
}

func (s *IamPolicySupplier) readPolicy(resource *iam.Policy) (cty.Value, error) {
	res, err := s.reader.ReadResource(
		terraform.ReadResourceArgs{
			Ty: s.SuppliedType(),
			ID: *resource.Arn,
		},
	)
	if err != nil {
		logrus.Warnf("Error reading iam policy %s[%s]: %+v", *resource.Arn, s.SuppliedType(), err)
		return cty.NilVal, err
	}

	return *res, nil
}
