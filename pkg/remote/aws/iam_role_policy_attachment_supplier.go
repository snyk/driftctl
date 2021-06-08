package aws

import (
	"fmt"

	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/cloudskiff/driftctl/pkg/terraform"

	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

type IamRolePolicyAttachmentSupplier struct {
	reader       terraform.ResourceReader
	deserializer *resource.Deserializer
	repo         repository.IAMRepository
	runner       *terraform.ParallelResourceReader
}

func NewIamRolePolicyAttachmentSupplier(provider *AWSTerraformProvider, deserializer *resource.Deserializer, repo repository.IAMRepository) *IamRolePolicyAttachmentSupplier {
	return &IamRolePolicyAttachmentSupplier{
		provider,
		deserializer,
		repo,
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s *IamRolePolicyAttachmentSupplier) SuppliedType() resource.ResourceType {
	return resourceaws.AwsIamRolePolicyAttachmentResourceType
}

func (s *IamRolePolicyAttachmentSupplier) Resources() ([]resource.Resource, error) {
	roles, err := s.repo.ListAllRoles()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationErrorWithType(err, s.SuppliedType(), resourceaws.AwsIamRoleResourceType)
	}
	policyAttachments, err := s.repo.ListAllRolePolicyAttachments(roles)
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, s.SuppliedType())
	}

	results := make([]cty.Value, 0)
	if len(policyAttachments) > 0 {
		for _, attachedPolicy := range policyAttachments {
			attached := *attachedPolicy
			s.runner.Run(func() (cty.Value, error) {
				return s.readRolePolicyAttachment(&attached)
			})
		}
		results, err = s.runner.Wait()
		if err != nil {
			return nil, err
		}
	}

	return s.deserializer.Deserialize(s.SuppliedType(), results)
}

func (s *IamRolePolicyAttachmentSupplier) readRolePolicyAttachment(attachedPol *repository.AttachedRolePolicy) (cty.Value, error) {
	res, err := s.reader.ReadResource(
		terraform.ReadResourceArgs{
			Ty: s.SuppliedType(),
			ID: fmt.Sprintf("%s-%s", *attachedPol.PolicyName, attachedPol.RoleName),
			Attributes: map[string]string{
				"role":       attachedPol.RoleName,
				"policy_arn": *attachedPol.PolicyArn,
			},
		},
	)

	if err != nil {
		logrus.Warnf("Error reading iam role policy attachment %s[%s]: %+v", attachedPol, s.SuppliedType(), err)
		return cty.NilVal, err
	}
	return *res, nil
}
